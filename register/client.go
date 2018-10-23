package register

import (
	"net"
	"github.com/leesper/holmes"
	"github.com/leesper/tao"
	"time"
	"fmt"
	"reflect"
)

type sTcpConfig struct {
	packageLengthOffset int    // 长度位移位
	packageBodyOffset   int    // length 包括 Header + Context + Body 的长度
	packageMaxLength    int    // 最大的长度
}

var (
	request Request
	tcpConfig sTcpConfig
)

// tcp标志位
const (
	FLAG_SYS_MSG     = 1 // 来自系统调用的消息请求
	FLAG_RESULT_MODE = 2 // 请求返回模式，表明这是一个RPC结果返回
	FLAG_FINISH      = 4 // 是否消息完成，用在消息返回模式里，表明RPC返回内容结束
	FLAG_TRANSPORT   = 8 // 转发浏览器RPC请求，表明这是一个来自浏览器的请求
)

func NewClient(address string) *tao.ClientConn {

	defer holmes.Start().Stop()

	if address == "" {
		holmes.Fatalln("缺少address")
	}
	c := doConnect(address)

	tao.Register(request.MessageNumber(), DeserializeRequest, nil)

	onConnect := tao.OnConnectOption(func(c tao.WriteCloser) bool {
		holmes.Infoln("on connect")
		return true
	})

	onError := tao.OnErrorOption(func(c tao.WriteCloser) {
		holmes.Infoln("on error")
	})

	onClose := tao.OnCloseOption(func(c tao.WriteCloser) {
		holmes.Infoln("on close")

		// 连接关闭 1秒后重连
		Logger.Error("RPC服务连接关闭，等待重新连接")
		time.Sleep(1 * time.Second)
		//cli.Connect()
	})

	onMessage := tao.OnMessageOption(func(msg tao.Message, c tao.WriteCloser) {
		ver     := msg.(Request).Ver
		flag    := msg.(Request).Flag
		body    := msg.(Request).Body
		header  := msg.(Request).Header
		context := msg.(Request).Context

		// 返回数据的模式
		if (flag & FLAG_RESULT_MODE) == FLAG_RESULT_MODE {
			holmes.Infoln("返回数据的模式", msg.(Request))
			rpcReceive(flag, header, body)
			return
		}

		//MyRpc.context.BaseContext.Set("receiveParam")
		socketSend(flag, ver, header, context, RpcHandle(body))
	})

	options := []tao.ServerOption{
		onConnect,
		onError,
		onClose,
		onMessage,
		//tao.ReconnectOption(),
		tao.CustomCodecOption(request),
	}

	return tao.NewClientConn(0, c, options...)
}

func doConnect(address string) net.Conn {
	c, err := net.Dial("tcp", address)
	if err != nil {
		holmes.Errorln(err)
		Logger.Error("RPC服务连接错误，等待重新连接" + err.Error())
		time.Sleep(1 * time.Second)
		return doConnect(address)
	}
	return c
}

func (reg *SRegister) Connect() {
	reg.Conn.Start()
	defer reg.Conn.Close()

	for {
		select {
		case <-time.After(6 * time.Second):
			go testRpcCall()

		case send := <-socketSendChan:
			if err := reg.Conn.Write(send); err != nil {
				holmes.Infoln("error", err)
			}
		}
	}
}


func testRpcCall() {
	time1 := time.Now().Unix()
	fmt.Println("rpc 请求", time1)
	args :=[]reflect.Value {reflect.ValueOf(time1)}
	result := RpcCall("test", args, "player", make(map[string]uint32))
	fmt.Println("rpc返回结果", result, time1)
}