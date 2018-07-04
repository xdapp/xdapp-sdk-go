package main

import (
	"server-register-go/register"
	"server-register-go/service"
)

/**
	 测试注册服务
 */
func main() {

	register.SetDebug(true)
	myReg := register.NewRegister()
	register.LoadService("", service.NewService(myReg))
	register.LoadService("sys", service.NewSysService(myReg))

	myReg.CreateClient()
}