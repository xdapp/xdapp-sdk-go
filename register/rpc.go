package register

import (
	"fmt"
	"github.com/hprose/hprose-golang/io"
	"github.com/hprose/hprose-golang/rpc"
	"reflect"
)

var (
	HproseService *rpc.TCPService		 // rpc 服务
	HproseContext *rpc.SocketContext	 // 上下文
)

func init() {
	HproseService = rpc.NewTCPService()
	HproseContext = new(rpc.SocketContext)
	HproseContext.InitServiceContext(HproseService)
}

// 屏蔽列表输出
func DoFunctionList() string {
	return "Fa{}z"
}

// 执行结果
func RpcHandle(data []byte) []byte {
	return HproseService.Handle(data, HproseContext)
}

// 已注册的rpc方法
func GetHproseAddedFunc() []string {
	return HproseService.MethodNames
}

// Simple 简单数据 https://github.com/hprose/hprose-golang/wiki/Hprose-%E6%9C%8D%E5%8A%A1%E5%99%A8
func AddFunction(name string, function interface{}) {
	HproseService.AddFunction(name, function, rpc.Options{Simple: true})
}

// 注册一个前端页面可访问的方法
func AddSysFunction(obj interface{}) {
	HproseService.AddInstanceMethods(obj, rpc.Options{NameSpace: "sys", Simple: true})
}

// 注册一个前端页面可访问的方法
func AddWebFunction(name string, function interface{}) {
	funcName := fmt.Sprintf("%s_%s", config.Name, name)
	HproseService.AddFunction(funcName, function, rpc.Options{Simple: true})
}

func AddWebInstanceMethods(obj interface{}, namespace string) {
	nsName := fmt.Sprintf("%s_%s", config.Name, namespace)
	HproseService.AddInstanceMethods(obj, rpc.Options{NameSpace: nsName, Simple: true})
}

func rpcEncode(name string, args []reflect.Value) []byte {
	writer := io.NewWriter(false)
	writer.WriteByte(io.TagCall)
	writer.WriteString(name)
	writer.Reset()
	writer.WriteSlice(args)
	writer.WriteByte(io.TagEnd)
	return writer.Bytes()
}

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