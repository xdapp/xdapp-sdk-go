package main

import (
	"server-register-go/register"
	"server-register-go/service"
)

/**
	 测试注册服务
 */
func main() {

	myReg := register.NewRegister()

	regInterface := service.RegisterInterFace(myReg)
	rpcService   := &register.RpcService{
		service.NewSysService(regInterface), service.NewService(regInterface)}
	register.LoadService(rpcService)

	myReg.CreateClient()
}