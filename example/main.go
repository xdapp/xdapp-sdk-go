package main

import (
	"server-register-go/register"
	"server-register-go/service"
)

/**
测试注册服务
 */
func main() {

	myReg, err := register.New(register.RegConfig{
		IsDebug: false,
	})

	if err != nil {
		panic(err.Error())
	}

	// 加载rpc 方法
	register.LoadService("sys", &service.Sys{myReg})

	// 增加扩展类
	register.LoadService("test", &service.Test{"test service"})

	// 增加单个方法
	register.MyRpc.AddFunction("hello", func() string {
		return "hello world"
	})

	myReg.CreateClient()
}