package register

import (
	"github.com/hprose/hprose-golang/rpc"
)

type sMyRpc struct {
	service  *rpc.TCPService			// rpc 服务
	context *rpc.SocketContext			// 上下文
}

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
	myRpc.service.AddFunction(name, function, option...)
}