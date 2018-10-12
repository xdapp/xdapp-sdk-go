package register

import (
	"reflect"
	"github.com/hprose/hprose-golang/rpc"
	"github.com/hprose/hprose-golang/io"
	"runtime"
	"time"
	"strings"
)

var (
	RpcService *rpc.TCPService		 // rpc 服务
	RpcContext *rpc.SocketContext	 // 上下文
	receiveBuffer  map[uint16][]byte // 接收的rpc请求数据
)

const (
	callTimeout   = 10
)

type SCallConfig struct {
	Timeout int
	AdminId uint32
	ServiceId uint32
	Namespace string
}

func init() {
	RpcService = rpc.NewTCPService()
	RpcContext = new(rpc.SocketContext)
	RpcContext.InitServiceContext(RpcService)
}

/**
屏蔽列表输出
*/
func DoFunctionList() string {
	return "Fa{}z"
}

/**
执行结果
*/
func RpcHandle(data []byte) []byte {
	return RpcService.Handle(data, RpcContext)
}

func AddFunction(name string, function interface{}, option ...rpc.Options) {
	RpcService.AddFunction(name, function, option...)
}

func AddInstanceMethods(obj interface{}, namespace string) {
	RpcService.AddInstanceMethods(obj, rpc.Options{NameSpace: namespace})
}

/*
# 将请求发送给RPC连接进程

# 标识   | 版本    | 长度    | 头信息       | 自定义内容    |  正文
# ------|--------|---------|------------|-------------|-------------
# Flag  | Ver    | Length  | Header     | Context      | Body
# 1     | 1      | 4       | 17         | 默认0，不定   | 不定
# C     | C      | N       |            |             |
#
#
# 其中 Header 部分包括
#
# AppId     | 服务ID      | rpc请求序号  | 管理员ID      | 自定义信息长度
# ----------|------------|------------|-------------|-----------------
# AppId     | ServiceId  | RequestId  | AdminId     | ContextLength
# 4         | 4          | 4          | 4           | 1
# N         | N          | N          | N           | C
*/
func (reg *SRegister) RpcCall(name string, args []reflect.Value, cfg SCallConfig) interface{} {

	if cfg.Timeout == 0 {
		cfg.Timeout = callTimeout
	}

	if cfg.Namespace != "" {
		nameSpace := strings.TrimSuffix(cfg.Namespace, "_") + "_"
		name = nameSpace + name
	}

	id         := uint32(1)
	custom     := Pack(uint16(0))
	contextLen := uint8(len(custom))
	rpcBody    := rpcEncode(name, args)

	version    := StrToByte(defaultVersion)
	length     := uint32(HEADER_LENGTH + len(custom) + len(rpcBody))
	request    := &SRequest{
		0,
		version,
		length,
		SHeader{
			0,
			cfg.ServiceId,
			id,
			cfg.AdminId,
			contextLen}}

	data := BytesCombine(Pack(request), custom, rpcBody)
	reg.Client.Send(data)

	runtime.Gosched()
	defer delete(rpcCallChan1, string(id))

	// 判断超时
	tick := time.After(callTimeout * time.Second)
	for {
		select {
		case <-tick:
			return "请求超时"
		case <-rpcCallChan1[string(id)]:
			return <-rpcCallChan1[string(id)]
		}
	}

}

func rpcMessage(id uint16, data []byte, finish bool)  {

	if finish == false {
		receiveBuffer[id] = data
	} else if receiveBuffer[id] != nil {
		data = BytesCombine(receiveBuffer[id], data)
		delete(receiveBuffer, id)
	}
	result, error := rpcDecode(data)

	if error != "" {
		Logger.Warn(error)
	} else {
		rpcCallChan1[string(id)]<-result
	}
}

/**
rpc序列化
 */
func rpcEncode(name string, args []reflect.Value) []byte {
	writer := io.NewWriter(false)
	writer.WriteByte(io.TagCall)
	writer.WriteString(name)
	writer.Reset()
	writer.WriteSlice(args)
	writer.WriteByte(io.TagEnd)
	return writer.Bytes()
}

/**
rpc反序列化
 */
func rpcDecode(data []byte) (interface{}, string) {
	reader := io.AcquireReader(data, false)
	defer io.ReleaseReader(reader)
	tag, _ := reader.ReadByte()
	switch tag {
	case io.TagResult:
		var e interface{}
		reader.Unserialize(&e)
		return e, ""
	case io.TagError:
		return nil, "RPC 系统调用 Agent 返回错误信息: " + reader.ReadString()
	default:
		return nil, "RPC 系统调用收到一个未定义的方法返回: " + string(tag) + reader.ReadString()
	}
}