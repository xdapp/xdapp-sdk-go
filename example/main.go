package main

import (
	"hub000.xindong.com/core-system/server-register-go/register"
	"hub000.xindong.com/core-system/server-register-go/service"
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