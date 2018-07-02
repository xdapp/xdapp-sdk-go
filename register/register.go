package register

import (
	"bytes"
)

var config = &tcpConfig {
	1,				// 包长开始位
	13,				// 1字节消息类型+4字节消息体长度+4字节用户id+4字节原消息fd+内容（id+data）
	0x200000}			// 最大包长度

/**
	创建一个连接客户端
 */
func (cli *Client) CreateServiceSocket() {

	cli.OnReceive(func(message []byte) {

		request := new(ReqestData)
		request.Unpack(bytes.NewReader(message))

		//myRpc.context.BaseContext.Set("receiveParam")

		// 执行rpc返回
		rpcData := myRpc.handle(request.Data, myRpc.context)

		rs := string(PackId(request.Id)) + string(rpcData)

		dataLen := len(rs);
		if dataLen < config.packageMaxLength {
			sendAnswer(cli, request.Flag | 4, request.Fd, string(rs))

		} else {
			for i := 0; i < dataLen; i += config.packageMaxLength {

				chunkLength := Min(config.packageMaxLength, dataLen - i)
				chunk := Substr(string(rs), i, chunkLength)

				flag := request.Flag
				if dataLen - i == chunkLength {
					flag |= 4
				}
				sendAnswer(cli, flag, request.Fd, chunk)
			}
		}
	})

	cli.Connect()
}

/**
	发送结果返回
 */
func sendAnswer(cli *Client, flag byte, fd uint32, data string) {
	response := &ResponseData{
		Flag: flag,
		Fd:   fd,
		Data: []byte(data),
	}
	response.Len = uint32(len(data))

	buf := new(bytes.Buffer)
	response.Pack(buf)
	cli.SendMessage(buf.Bytes())
}

/**
	初始化tcp client
 */
func (reg *RegisterData) CreateClient() {
	reg.MyClient.CreateServiceSocket()
}