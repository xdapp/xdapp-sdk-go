package register

import (
	"net"
	"time"
	"os"
	"encoding/binary"
	"bytes"
	"bufio"
)

type TcpEvent interface {
	OnConnect()
	OnClose()
	OnError()
	OnReceive()
}

type Client struct {
	tcpConfig
	Addr   *net.TCPAddr
	Conn   *net.TCPConn

	// 调用时重写回调方法
	onCloseCallback   func()
	onErrorCallback  func(err error)
	onReceiveCallback func(message []byte)
	onSpiltCallback func(data []byte, atEOF bool) (advance int, token []byte, err error)
}

type tcpConfig struct {
	packageLengthOffset		int		// 长度位移位
	packageBodyOffset		int		// 1字节消息类型+4字节消息体长度+4字节用户id+4字节原消息fd+内容（id+data）
	packageMaxLength		int		// 最大的长度
}

const Default_Max_Len  = 0x200020

/**
	连接
 */
func (cli *Client) OnConnect() {
	conn, err := net.DialTCP("tcp", nil, cli.Addr)
	//conn, err := net.DialTimeout("tcp", nil, cli.Addr, 2*time.Second)

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
func (cli *Client) OnError(callback func( err error)) {
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
func (cli *Client) SendMessage(data []byte) {
	_, err := cli.Conn.Write(data)
	if err != nil {
		Debug("发送失败" + err.Error())
	}
}

/**
	连接
 */
func (cli *Client) Connect() {

	cli.OnConnect()

	Info("##############", cli.Conn)

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
		Error("scanner", err.Error())
		//cli.onErrorCallback(err)
		return
	}

	defer cli.Close(true)
}

/**
	关闭client
 */
func (cli *Client) Close(regSuccess bool) {
	cli.Conn.Close()
	defer Debug("status", regSuccess)

	if regSuccess {
		cli.onCloseCallback()
	}
}

/**
	创建客户端
 */
func NewClient(address string, config tcpConfig) *Client {

	cli := Client{}

	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	cli.Addr = tcpAddr
	cli.tcpConfig = config

	cli.OnReceive(func(message []byte) {
		data := make([]byte, Default_Max_Len)
		n, _ := cli.Conn.Read(data)
		Debug(data[:n])
	})

	cli.OnClose(func() {
		// 连接关闭 1秒后重连
		Debug("RPC服务连接关闭，等待重新连接")
		time.Sleep(1 * time.Second)
		cli.Connect()
	})

	cli.OnError(func(err error) {
		// 连接失败 1秒后重连
		Debug("RPC服务连接错误，等待重新连接" + err.Error())
		time.Sleep(1 * time.Second)
		cli.Connect()
	})

	cli.OnSplit(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		//if !atEOF && data[0] == 'V' {
		if !atEOF {

			// 超过最大长度2m
			if len(data) > cli.packageMaxLength {
				return
			}

			// 最小13字节 1字节消息类型+4字节消息体长度+4字节用户id+4字节原消息fd+内容（id+data）
			if len(data) < cli.packageBodyOffset {
				return
			}
			Length := uint32(0)
			binary.Read(bytes.NewReader(data[cli.packageLengthOffset : cli.packageLengthOffset + 4]), binary.BigEndian, &Length)

			// 读取到的数据正文长度 + 13字节 不超过读到的原始数据长度 则拆包
			if int(Length) + cli.packageLengthOffset <= len(data) {
				return int(Length) + cli.packageBodyOffset, data[:int(Length) + cli.packageBodyOffset], nil
			}
		}
		return
	})

	return &cli
}