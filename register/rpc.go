package register

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hprose/hprose-golang/io"
	"github.com/hprose/hprose-golang/rpc"
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

func GetCurrentAdminId() uint {
	return hproseContext.GetUInt("adminId")
}

func GetCurrentAppId() uint {
	return hproseContext.GetUInt("appId")
}

func GetCurrentServiceId() uint {
	return hproseContext.GetUInt("serviceId")
}

func GetCurrentRequestId() uint {
	return hproseContext.GetUInt("requestId")
}

// 执行结果
func RpcHandle(header Header, data []byte) []byte {
	hproseContext = new(rpc.SocketContext)
	hproseContext.InitServiceContext(hproseService)
	hproseContext.SetUInt("appId", uint(header.AppId))
	hproseContext.SetUInt("serviceId", uint(header.ServiceId))
	hproseContext.SetUInt("requestId", uint(header.RequestId))
	hproseContext.SetUInt("adminId", uint(header.AdminId))
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
	hproseService.AddFilter(filter...)
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

func AddBeforeFilterHandler(handle ...rpc.FilterHandler) {
	hproseService.AddBeforeFilterHandler(handle...)
}

func rpcEncode(name string, args []reflect.Value) []byte {
	w := io.NewWriter(false)
	w.WriteByte(io.TagCall)
	w.WriteString(name)
	w.Reset()
	w.WriteSlice(args)
	w.WriteByte(io.TagEnd)
	return w.Bytes()
}

func rpcDecode(data []byte) (interface{}, error) {
	r := io.AcquireReader(data, false)
	defer io.ReleaseReader(r)
	tag, _ := r.ReadByte()
	switch tag {
	case io.TagResult:
		var e interface{}
		r.Unserialize(&e)
		return e, nil
	case io.TagError:
		return nil, errors.New("RPC 系统调用 Agent 返回错误信息: " + r.ReadString())
	default:
		return nil, errors.New("RPC 系统调用收到一个未定义的方法返回: " + string(tag) + r.ReadString())
	}
}
