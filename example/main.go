package main

import (
	"server-register-go/register"
	"server-register-go/service"
)

/**
	 测试注册服务
 */
func main() {

	myReg := register.NewRegister(register.RegConfig{
		IsDebug: false,
	})

	// 加载rpc 方法
	register.LoadService("", service.NewService(myReg))
	register.LoadService("sys", service.NewSysService(myReg))

	// 增加单个方法
	register.MyRpc.AddFunction("test", func() string {
		return "just test"
	})

	myReg.CreateClient()
}