package register

import (
	"github.com/hprose/hprose-golang/rpc"
)

type MyRpc struct {
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
func NewMyRpc() *MyRpc {
	service := rpc.NewTCPService()
	context := new(rpc.SocketContext)
	context.InitServiceContext(service)
	return &MyRpc{service: service, context:context}
}

/**
	执行结果
 */
func (mrpc *MyRpc) handle(data []byte, context rpc.Context) []byte {
	return mrpc.service.Handle(data, context)
}