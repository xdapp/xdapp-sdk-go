package main

import (
	"server-register-go/register"
	"server-register-go/service"
	"github.com/leesper/holmes"
	"time"
)

/**
测试注册服务
*/
func main() {
	//defer holmes.Start(holmes.LogFilePath("./log"), holmes.EveryHour, holmes.AlsoStdout).Stop()
	defer holmes.Start().Stop()

	myReg, err := register.New(register.RegConfig{IsDebug: true})
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
	register.PrintRpcAddFunctions()

	myReg.Conn.Start()
	defer myReg.Conn.Close()

	// 测试请求rpc
	for {
		select {
		case <-time.After(6 * time.Second):
			go register.TestRpcCall()
		}
	}
}