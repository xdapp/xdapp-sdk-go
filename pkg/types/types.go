package types

const (
	RPCVersion       = 1       // 默认RPC版本
	PackageMaxLength = 0x21000 // 最大包长度
	LogFileName   = "test.log"  // log文件名

	RpcCallWorkId   = 0  // rpc workId (PHP版本对应进程id)
	RpcCallTimeout  = 10 // rpc 请求超时时间
	RpcClearBufTime = 30 // rpc 清理数据时间

	ProtocolFlagSysMsg     = 1 // 来自系统调用的消息请求
	ProtocolFlagResultMode = 2 // 请求返回模式，表明这是一个RPC结果返回
	ProtocolFlagFinish     = 4 // 是否消息完成，用在消息返回模式里，表明RPC返回内容结束
	ProtocolFlagTransport  = 8 // 转发浏览器RPC请求，表明这是一个来自浏览器的请求


	//  标识   | 版本    | 长度    | 头信息       | 自定义内容    |  正文
	//  ------|--------|---------|------------|-------------|-------------
	//  Flag  | Ver    | Length  | Header     | Context      | Body
	//  1     | 1      | 4       | 17         | 默认0，不定   | 不定
	//  C     | C      | N       |            |             |
	//
	//
	//  其中 Header 部分包括
	//
	//  服务ID     | rpc请求序号  | 管理员ID      | 自定义信息长度
	//  ----------|-------------|-------------|-----------------
	//  ServiceId | RequestId   | AdminId     | ContextLength
	//  4         | 4           | 4           | 1
	//  N         | N           | N           | C
	ProtocolPrefixLength    = 6                           // Flag 1字节、 Ver 1字节、 Length 4字节
	ProtocolHeaderLength    = 17                          // 默认消息头长度, 不包括 PrefixLength
	ProtocolContextOffset   = ProtocolPrefixLength + ProtocolHeaderLength // 自定义上下文内容所在位置，   23
	ProtocolSendChunkLength = 0x200000                    // 单次发送的包大小
)