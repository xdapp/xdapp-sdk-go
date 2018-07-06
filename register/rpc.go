package register

import (
	"github.com/hprose/hprose-golang/rpc"
	"reflect"
	"strings"
)

type sMyRpc struct {
	service  *rpc.TCPService			// rpc 服务
	context *rpc.SocketContext			// 上下文
}

/**
	成功服务列表
 */
var sucRpcFunc = make(map[string]string)

/**
	屏蔽列表输出
  */
func DoFunctionList() string {
	return "Fa{}z"
}

/**
	初始化
 */
func NewMyRpc() *sMyRpc {
	service := rpc.NewTCPService()
	context := new(rpc.SocketContext)
	context.InitServiceContext(service)
	return &sMyRpc{service: service, context:context}
}

/**
	执行结果
 */
func (myRpc *sMyRpc) handle(data []byte, context rpc.Context) []byte {
	return myRpc.service.Handle(data, context)
}

/**
	注册方法
 */
func (myRpc *sMyRpc) AddFunction(name string, function interface{}, option ...rpc.Options) {

	if sucRpcFunc[name] != ""  {
		MyLog.Error("RPC服务已经存在 " + name + ", 已忽略 ")
	}
	MyLog.Debug("增加RPC方法： " + name)

	myRpc.service.AddFunction(name, function, option...)
	sucRpcFunc[name] = name + "()"
}

/**
	加载service 作为receiver的可执行的所有方法
  */
func LoadService(prefix string, service interface{}) {

	t := reflect.TypeOf(service)
	v := reflect.ValueOf(service)

	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		mv := v.MethodByName(m.Name) 	//获取对应的方法

		if !mv.IsValid() {            	//判断方法是否存在
			MyLog.Error(m.Name + "方法不存在！")
			continue
		}

		// 注册rpc方法
		rpcName := strings.ToLower(m.Name)
		if prefix != "" {
			rpcName = prefix + "_" + strings.ToLower(m.Name)
		}
		MyRpc.AddFunction(rpcName, mv)
	}
}

/**
	打印已注册rpc服务列表
 */
func debugSuccessService() {
	if  sucRpcFunc != nil {
		MyLog.Debug("已注册的rpc服务列表: " + Implode(",", sucRpcFunc))
	}
}