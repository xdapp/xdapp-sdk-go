package main

import (
	"server-register-go/register"
	"server-register-go/service"
	"time"
)

/**
测试注册服务
*/
func main() {
	sReg, err := register.New(register.RegConfig{IsDebug: true})
	if err != nil {
		panic(err.Error())
	}

	// 加载rpc 方法
	register.AddInstanceMethods(&service.Sys{sReg}, "sys")
	register.AddInstanceMethods(&service.Test{"test service"}, "test")

	// 增加单个方法
	register.AddFunction("hello", func() string {
		return "hello world"
	})
	register.PrintRpcAddFunctions()

	sReg.Conn.Start()
	defer sReg.Conn.Close()

	for {
		select {
		case <-time.After(6 * time.Second):
			go register.TestRpcPing()
		}
	}
}