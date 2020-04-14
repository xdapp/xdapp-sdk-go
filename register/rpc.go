package register

import (
	"fmt"
	"github.com/hprose/hprose-golang/io"
	"github.com/hprose/hprose-golang/rpc"
	"reflect"
)

var (
	hproseService *rpc.TCPService
	hproseContext *rpc.SocketContext
)

func init() {
	hproseService = rpc.NewTCPService()
}

// 屏蔽列表输出
func DoFunctionList() string {
	return "Fa{}z"
}

// 获取当前上下文
func GetCurrentContext() *rpc.SocketContext {
	return hproseContext
}

func GetCurrentAdminId() uint32 {
	id := hproseContext.GetInterface("adminId")
	res, _ := id.(uint32)
	return res
}

func GetCurrentAppId() uint32 {
	id := hproseContext.GetInterface("appId")
	res, _ := id.(uint32)
	return res
}

func GetCurrentServiceId() uint32 {
	id := hproseContext.GetInterface("serviceId")
	res, _ := id.(uint32)
	return res
}

func GetCurrentRequestId() uint32 {
	id := hproseContext.GetInterface("requestId")
	res, _ := id.(uint32)
	return res
}

// 执行结果
func RpcHandle(header Header, data []byte) []byte {
	hproseContext = new(rpc.SocketContext)
	hproseContext.InitServiceContext(hproseService)
	hproseContext.Set("appId", header.AppId)
	hproseContext.Set("serviceId", header.ServiceId)
	hproseContext.Set("requestId", header.RequestId)
	hproseContext.Set("adminId", header.AdminId)
	return hproseService.Handle(data, hproseContext)
}

// 已注册的rpc方法
func GetHproseAddedFunc() []string {
	return hproseService.MethodNames
}

// Simple 简单数据 https://github.com/hprose/hprose-golang/wiki/Hprose-%E6%9C%8D%E5%8A%A1%E5%99%A8
func AddFunction(name string, function interface{}) {
	hproseService.AddFunction(name, function, rpc.Options{Simple: true})
}

// 注册一个前端页面可访问的方法
func AddSysFunction(obj interface{}) {
	hproseService.AddInstanceMethods(obj, rpc.Options{NameSpace: "sys", Simple: true})
}

// 增加过滤器
func AddFilter(filter ...rpc.Filter) {
	hproseService.AddFilter(filter ...)
}

// 注册一个前端页面可访问的方法
func AddWebFunction(name string, function interface{}) {
	funcName := fmt.Sprintf("%s_%s", config.Name, name)
	hproseService.AddFunction(funcName, function, rpc.Options{Simple: true})
}

func AddWebInstanceMethods(obj interface{}, namespace string) {
	nsName := config.Name
	if namespace != "" {
		nsName = fmt.Sprintf("%s_%s", config.Name, namespace)
	}
	hproseService.AddInstanceMethods(obj, rpc.Options{NameSpace: nsName, Simple: true})
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