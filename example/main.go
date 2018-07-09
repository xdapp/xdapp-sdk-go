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
	register.LoadService("sys", service.NewSysService(myReg))

	// 增加扩展类
	register.LoadService("test", service.NewTestService("test service"))

	// 增加单个方法
	register.MyRpc.AddFunction("hello", func() string {
		return "hello world"
	})

	myReg.CreateClient()
}