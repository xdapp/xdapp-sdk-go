package main

import (
	"server-register-go/register"
	"server-register-go/service"
	"fmt"
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
	register.AddInstanceMethods(&service.Sys{myReg}, "sys")
	register.AddInstanceMethods(&service.Test{"test service"}, "test")

	// 增加单个方法
	register.AddFunction("hello", func() string {
		return "hello world"
	})

	fmt.Println(myReg.GetFunctions())

	myReg.Client.Connect()
}
