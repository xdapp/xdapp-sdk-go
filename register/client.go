package register

import (
	"net"
	"time"
	"github.com/leesper/tao"
)

var (
	request Request
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
		Logger.Error("缺少tcp host")
	}

	c := doConnect(host)

	onConnect := tao.OnConnectOption(func(c tao.WriteCloser) bool {
		return true
	})

	onError := tao.OnErrorOption(func(c tao.WriteCloser) {
	})

	onClose := tao.OnCloseOption(func(c tao.WriteCloser) {
		// 连接关闭 1秒后重连
		Logger.Error("RPC服务连接关闭，等待重新连接")
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
		//RpcContext.BaseContext.Set("receiveParam", msg.(Request))
		transportRpcRequest(flag, ver, header, context, RpcHandle(body))
	})

	options := []tao.ServerOption{
		onConnect,
		onError,
		onClose,
		onMessage,
		tao.ReconnectOption(),
		tao.CustomCodecOption(request),
	}

	tao.Register(request.MessageNumber(), unserialize, nil)
	Conn = tao.NewClientConn(0, c, options...)
	return Conn
}

func doConnect(host string) net.Conn {
	c, err := net.Dial("tcp", host)
	if err != nil {
		Logger.Error("RPC服务连接错误，等待重新连接" + err.Error())
		time.Sleep(1 * time.Second)
		return doConnect(host)
	}
	return c
}

func (reg *SRegister) Connect() {
	reg.Conn.Start()
	outputAddedFunctions()
	<- startChan
}