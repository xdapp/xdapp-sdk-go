package register

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"fmt"
	"time"
)

type TcpEvent interface {
	OnConnect()
	OnClose()
	OnError()
	OnReceive()
}

type Client struct {
	tcpConfig
	host string
	Conn *net.TCPConn

	// 调用时重写回调方法
	onCloseCallback   func()
	onErrorCallback   func(err error)
	onReceiveCallback func(message []byte)
	onSpiltCallback   func(data []byte, atEOF bool) (advance int, token []byte, err error)
}

// 标识   | 版本    | 长度    | 头信息       | 自定义上下文  |  正文
// ------|--------|---------|------------|-------------|-------------
// Flag  | Ver    | Length  | Header     | Context     | Body
// 1     | 1      | 4       | 17         | 默认0不定    | 不定
// C     | C      | N       |            |             |
// length 包括 Header + Context + Body 的长度

type tcpConfig struct {
	packageLengthOffset int // 长度位移位
	packageBodyOffset   int // length 包括 Header + Context + Body 的长度
	packageMaxLength    int // 最大的长度
}

// tcp标志位
const (
	FLAG_SYS_MSG     = 1 // 来自系统调用的消息请求
	FLAG_RESULT_MODE = 2 // 请求返回模式，表明这是一个RPC结果返回
	FLAG_FINISH      = 4 // 是否消息完成，用在消息返回模式里，表明RPC返回内容结束
	FLAG_TRANSPORT   = 8 // 转发浏览器RPC请求，表明这是一个来自浏览器的请求
)

/**
连接
*/
func (cli *Client) OnConnect() {

	Logger.Debug("tcp连接ip地址：" + cli.host)

	tcpAddr, err := net.ResolveTCPAddr("tcp", cli.host)
	if err != nil {
		log.Fatal("ResolveTCPAddr failed:", err.Error())
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	//conn, err := net.DialTimeout("tcp", cli.host, 1*time.Second)
	if err != nil {
		cli.onErrorCallback(err)
	}
	cli.Conn = conn
}

/**
收到消息
*/
func (cli *Client) OnReceive(callback func(message []byte)) {
	cli.onReceiveCallback = callback
}

/**
关闭
*/
func (cli *Client) OnClose(callback func()) {
	cli.onCloseCallback = callback
}

/**
错误处理
*/
func (cli *Client) OnError(callback func(err error)) {
	cli.onErrorCallback = callback
}

/**
分割信息
*/
func (cli *Client) OnSplit(callback func(data []byte, atEOF bool) (advance int, token []byte, err error)) {
	cli.onSpiltCallback = callback
}

/**
发送消息
*/
func (cli *Client) Send(data []byte) {
	_, err := cli.Conn.Write(data)
	if err != nil {
		Logger.Debug("发送失败" + err.Error())
	}
}

/**
连接
*/
func (cli *Client) Connect() {

	cli.OnConnect()
	defer cli.Close(true)
	if cli.Conn == nil {
		return
	}

	/**
	解决粘包的问题
	*/
	scanner := bufio.NewScanner(cli.Conn)
	scanner.Split(cli.onSpiltCallback)

	for scanner.Scan() {
		cli.onReceiveCallback(scanner.Bytes())
	}

	if err := scanner.Err(); err != nil {
		Logger.Error("scanner", err.Error())
		//cli.onErrorCallback(err)
		return
	}
}

/**
关闭client
*/
func (cli *Client) Close(regSuccess bool) {
	cli.Conn.Close()
	defer Logger.Debug("status", regSuccess)

	if regSuccess {
		cli.onCloseCallback()
	}
}

/**
创建客户端
*/
func NewClient(host string, tcpConf tcpConfig) *Client {

	cli := Client{host: host, tcpConfig: tcpConf}

	cli.OnSplit(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		//if !atEOF && data[0] == 'V' {
		if !atEOF {

			// 超过最大长度2m
			if len(data) > cli.packageMaxLength {
				return
			}
			if len(data) < cli.packageBodyOffset {
				return
			}

			Length := uint32(0)
			lengthStart := cli.packageLengthOffset
			lengthEnd := cli.packageBodyOffset
			binary.Read(bytes.NewReader(data[lengthStart:lengthEnd]), binary.BigEndian, &Length)

			// 拆包
			if int(Length) + cli.packageLengthOffset <= len(data) {
				bodyOffsetEnd := int(Length) + cli.packageBodyOffset
				return bodyOffsetEnd, data[:bodyOffsetEnd], nil
			}
		}
		return
	})

	cli.OnReceive(func(message []byte) {
		fmt.Println(string(message))

		ver  := getVer(message)
		flag := getFlag(message)

		if ver != 1 {
			Logger.Error("消息版本错误",  ver)
			return
		}

		// 返回数据的模式
		if (flag & FLAG_RESULT_MODE) == FLAG_RESULT_MODE {
			finish := (flag & FLAG_FINISH) == FLAG_FINISH
			//workerId := 1
			fmt.Println(finish)
			return
		}

		request     := NewRequest(message[:CONTEXT_OFFSET])
		rpcData     := message[(CONTEXT_OFFSET + request.ContextLength):]
		headContext := message[PREFIX_LENGTH:(PREFIX_LENGTH + HEADER_LENGTH + request.ContextLength)]

		//MyRpc.context.BaseContext.Set("receiveParam")
		rpcResponse := RpcHandle(rpcData)

		cli.sendSocket(rpcResponse, headContext, ver, flag)
	})

	cli.OnClose(func() {
		// 连接关闭 1秒后重连
		Logger.Error("RPC服务连接关闭，等待重新连接")
		time.Sleep(1 * time.Second)
		cli.Connect()
	})

	cli.OnError(func(err error) {
		// 连接失败 1秒后重连
		Logger.Error("RPC服务连接错误，等待重新连接" + err.Error())
		time.Sleep(1 * time.Second)
		cli.Connect()
	})

	return &cli
}

func (cli *Client)sendSocket(data[]byte, headContext []byte, ver byte, flag byte) {

	flag = flag | FLAG_RESULT_MODE
	dataLength := len(data)
	headerAndContextLen := len(headContext)

	if dataLength < 0x200000 {
		response := &SResponse{
			uint8(flag | FLAG_FINISH), ver, uint32(headerAndContextLen + dataLength),
		}
		sendData := BytesCombine(response.Pack(), headContext, data)
		cli.Send(sendData)
		return
	}

	// 大于 拆包分段发送
	for i := 0; i < dataLength; i += 0x200000 {
		chunkLen := Min(0x200000, dataLength - i)
		chunk := data[i:chunkLen]

		if dataLength - i == chunkLen {
			flag |= FLAG_FINISH
		}
		response := &SResponse{
			uint8(flag | FLAG_FINISH), ver, uint32(headerAndContextLen + dataLength),
		}
		sendData := BytesCombine(response.Pack(), headContext, []byte(chunk))
		cli.Send(sendData)
	}
}