package register

import (
	"crypto/tls"
	"net"
	"time"
	"github.com/leesper/tao"
)

var (
	request Request
	requestId *tao.AtomicInt64	// 请求id 原子递增
)

// tcp标志位
const (
	FLAG_SYS_MSG     = 1 // 来自系统调用的消息请求
	FLAG_RESULT_MODE = 2 // 请求返回模式，表明这是一个RPC结果返回
	FLAG_FINISH      = 4 // 是否消息完成，用在消息返回模式里，表明RPC返回内容结束
	FLAG_TRANSPORT   = 8 // 转发浏览器RPC请求，表明这是一个来自浏览器的请求
)

func NewClient(host string) *tao.ClientConn {
	if host == "" {
		Logger.Warn("缺少tcp host")
	}

	c := doConnect(host)
	onConnect := tao.OnConnectOption(func(c tao.WriteCloser) bool {
		return true
	})

	onError := tao.OnErrorOption(func(c tao.WriteCloser) {
	})

	// 连接关闭 1秒后重连
	onClose := tao.OnCloseOption(func(c tao.WriteCloser) {
		Logger.Debug("RPC服务连接关闭，等待重新连接")
		doConnect(host)
	})

	onMessage := tao.OnMessageOption(func(msg tao.Message, c tao.WriteCloser) {
		ver     := msg.(Request).Ver
		flag    := msg.(Request).Flag
		body    := msg.(Request).Body
		header  := msg.(Request).Header
		context := msg.(Request).Context

		// 返回数据的模式
		if (flag & FLAG_RESULT_MODE) == FLAG_RESULT_MODE {
			sendRpcReceive(flag, header, body)
			return
		}
		transportRpcRequest(flag, ver, header, context, RpcHandle(body))
	})

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}
	requestId = tao.NewAtomicInt64(0)
	options := []tao.ServerOption{
		onConnect,
		onError,
		onClose,
		onMessage,
		tao.TLSCredsOption(tlsConf),
		tao.ReconnectOption(),
		tao.CustomCodecOption(request),
	}

	tao.Register(request.MessageNumber(), Unserialize, nil)

	return tao.NewClientConn(0, c, options...)
}

func doConnect(host string) net.Conn {

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	c, err := tls.Dial("tcp", host, conf)

	if err != nil {
		Logger.Warn("RPC服务连接错误，等待重新连接" + err.Error())
		time.Sleep(2 * time.Second)
		return doConnect(host)
	}
	return c
}