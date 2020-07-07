package register

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/leesper/tao"
	"github.com/xdapp/xdapp-sdk-go/pkg/types"
)

var (
	request        Request
	requestId      *tao.AtomicInt64 		// 请求id 原子递增
	receiveBuff    *safeReceiveBuff			// 收到缓冲区
	receiveChanMap = make(map[string]chan interface{})
)

type safeReceiveBuff struct {
	mu     sync.Mutex
	bufMap map[string][]byte
}
var (
	tlsConf = &tls.Config{ InsecureSkipVerify: true }
)

func (reg *register) NewClient(host string, port int, ssl bool) *tao.ClientConn {
	if host == "" {
		err := types.ErrTCPRequireHost
		reg.lg.Error(err.Error())
	}

	onMessage := tao.OnMessageOption(func(msg tao.Message, c tao.WriteCloser) {
		req, ok := msg.(Request); if !ok {
			err := types.ErrTCPParseRequest
			reg.lg.Error(err.Error())
		}

		ver := req.Ver
		flag := req.Flag
		body := req.Body
		header := req.Header
		context := req.Context

		reg.lg.Debug("[tcp] 收到原始请求信息: " + string(body[:]))

		// 返回数据的模式
		if (flag & types.ProtocolFlagResultMode) == types.ProtocolFlagResultMode {
			err := reg.conn.sendRpcReceive(flag, header, body)
			if err != nil {
				err = fmt.Errorf("[rpc返回] 发送数据异常 %w", err)
				reg.lg.Error(err.Error())
			}
			return
		}
		go func() {
			rs := reg.RpcHandle(header, body)
			err := transportRpcRequest(c, flag, ver, header, context, rs)
			if err != nil {
				err = fmt.Errorf("[rpc转发] 发送数据异常 %w", err)
				reg.lg.Error(err.Error())
			}
		}()
	})

	requestId = tao.NewAtomicInt64(0)
	options := []tao.ServerOption{
		tao.OnConnectOption(func(c tao.WriteCloser) bool { return true }),
		tao.OnErrorOption(func(c tao.WriteCloser) {}),
		// 连接关闭 1秒后重连
		tao.OnCloseOption(func(c tao.WriteCloser) {
			reg.lg.Debug(types.DebugRetryTCPMessage)
			reg.doConnect(host, port, ssl)
		}),
		onMessage,
		tao.ReconnectOption(),
		tao.CustomCodecOption(request),
	}

	if ssl {
		options = append(options, tao.TLSCredsOption(tlsConf))
	}

	tao.Register(request.MessageNumber(), unSerialize, nil)

	return tao.NewClientConn(0, reg.doConnect(host, port, ssl), options...)
}

func (reg *register) doConnect(host string, port int, ssl bool) net.Conn {
	address := fmt.Sprintf("%s:%d", host, port)
	if ssl {
		c, err := tls.Dial("tcp", address, tlsConf)
		if err != nil {
			reg.lg.Warn("RPC服务连接错误，等待重新连接" + err.Error())
			time.Sleep(2 * time.Second)
			return reg.doConnect(host, port, ssl)
		}
		return c
	} else {
		c, err := net.Dial("tcp", address)
		if err != nil {
			reg.lg.Warn("RPC服务连接错误，等待重新连接" + err.Error())
			time.Sleep(2 * time.Second)
			return reg.doConnect(host, port, ssl)
		}
		return c
	}
}

// rpc 请求返回
func (c *clientConn) sendRpcReceive(flag byte, header Header, body []byte) error {
	reqIdInt := uint64(header.RequestId)
	reqId := strconv.FormatUint(reqIdInt, 10)

	finish := (flag & types.ProtocolFlagFinish) ==
		types.ProtocolFlagFinish

	if finish == false {
		receiveBuff.mu.Lock()
		receiveBuff.bufMap[reqId] = body
		receiveBuff.mu.Unlock()

		// 30秒后清理数据
		d := time.Now().Add(types.RpcClearBufTime*time.Second)
		c.RunAt(d, func(i time.Time, closer tao.WriteCloser) {
			receiveBuff.mu.Lock()
			delete(receiveBuff.bufMap, reqId)
			receiveBuff.mu.Unlock()
		})
		return nil
	} else if receiveBuff.bufMap[reqId] != nil {
		var b bytes.Buffer
		b.Write(receiveBuff.bufMap[reqId])
		b.Write(body)
		body = b.Bytes()

		receiveBuff.mu.Lock()
		delete(receiveBuff.bufMap, reqId)
		receiveBuff.mu.Unlock()
	}

	resp, err := rpcDecode(body)
	if err != nil {
		return err
	}
	receiveChanMap[reqId] = make(chan interface{})
	receiveChanMap[reqId] <- resp
	return nil
}
