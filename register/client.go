package register

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/leesper/tao"
	"github.com/xdapp/xdapp-sdk-go/pkg/types"
)

var (
	request        Request
	requestId      *tao.AtomicInt64 // 请求id 原子递增
	receiveBuff    *safeReceiveBuff
	receiveChanMap = make(map[string]chan interface{})
)

type safeReceiveBuff struct {
	bufMap map[string][]byte
	mu     sync.Mutex
}

func init() {
	receiveBuff = &safeReceiveBuff{}
}

func NewClient(host string, port int, ssl bool) *tao.ClientConn {
	if host == "" {
		Logger.Error("[tcp] 缺少host")
	}

	c := doConnect(host, port, ssl)
	onConnect := tao.OnConnectOption(func(c tao.WriteCloser) bool {
		return true
	})

	onError := tao.OnErrorOption(func(c tao.WriteCloser) {})

	// 连接关闭 1秒后重连
	onClose := tao.OnCloseOption(func(c tao.WriteCloser) {
		Logger.Debug("[tcp] RPC服务连接关闭，等待重新连接")
		doConnect(host, port, ssl)
	})

	onMessage := tao.OnMessageOption(func(msg tao.Message, c tao.WriteCloser) {
		req, ok := msg.(Request); if !ok {
			Logger.Error("[tcp] 解析数据格式异常")
		}

		ver := req.Ver
		flag := req.Flag
		body := req.Body
		header := req.Header
		context := req.Context

		// 返回数据的模式
		if (flag & types.ProtocolFlagResultMode) == types.ProtocolFlagResultMode {
			err := sendRpcReceive(flag, header, body)
			if err != nil {
				Logger.Error(err)
			}
			return
		}
		go func() {
			err := transportRpcRequest(c, flag, ver, header, context, RpcHandle(header, body))
			if err != nil {
				Logger.Error(err)
			}
		}()
	})

	requestId = tao.NewAtomicInt64(0)
	options := []tao.ServerOption{
		onConnect,
		onError,
		onClose,
		onMessage,
		tao.ReconnectOption(),
		tao.CustomCodecOption(request),
	}

	if ssl {
		tlsConf := &tls.Config{
			InsecureSkipVerify: true,
		}
		options = append(options, tao.TLSCredsOption(tlsConf))
	}

	tao.Register(request.MessageNumber(), unSerialize, nil)

	return tao.NewClientConn(0, c, options...)
}

func doConnect(host string, port int, ssl bool) net.Conn {
	address := fmt.Sprintf("%s:%s", host, IntToStr(port))
	if ssl {
		c, err := tls.Dial("tcp", address, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			Logger.Warn("RPC服务连接错误，等待重新连接" + err.Error())
			time.Sleep(2 * time.Second)
			return doConnect(host, port, ssl)
		}
		return c
	} else {
		c, err := net.Dial("tcp", address)
		if err != nil {
			Logger.Warn("RPC服务连接错误，等待重新连接" + err.Error())
			time.Sleep(2 * time.Second)
			return doConnect(host, port, ssl)
		}
		return c
	}
}

// rpc 请求返回
func sendRpcReceive(flag byte, header Header, body []byte) error {
	reqId := IntToStr(header.RequestId)
	finish := (flag & types.ProtocolFlagFinish) == types.ProtocolFlagFinish

	if finish == false {
		receiveBuff.mu.Lock()
		receiveBuff.bufMap[reqId] = body
		receiveBuff.mu.Unlock()

		// 30秒后清理数据
		d := time.Now().Add(types.RpcClearBufTime*time.Second)
		Conn.RunAt(d, func(i time.Time, closer tao.WriteCloser) {
			receiveBuff.mu.Lock()
			delete(receiveBuff.bufMap, reqId)
			receiveBuff.mu.Unlock()
		})
		return nil
	} else if receiveBuff.bufMap[reqId] != nil {
		body = BytesCombine(receiveBuff.bufMap[reqId], body)
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
