package main

import (
	"time"
	"reflect"
	"fmt"
	"server-register-go/register"
	"server-register-go/service"
	"github.com/hprose/hprose-golang/rpc"
)

/**
测试注册服务
*/
func main() {
	conf := register.Config{
		App: "demo",
		Name: "name",
		SSl: false,
		Key: "aaaaaaaaaa",
		Host: "172.26.128.162:8900",
		IsDebug: false,
	}
	sReg, err := register.New(conf)
	if err != nil {
		panic(err.Error())
	}

	// 加载rpc 方法
	hproseService := register.HproseService
	hproseService.AddInstanceMethods(&service.Sys{sReg}, rpc.Options{NameSpace: "sys"})
	hproseService.AddInstanceMethods(&service.Test{"test service"}, rpc.Options{NameSpace: "test"})

	hproseService.AddFunction("hello", func() string {
		return "hello world"
	})
	register.PrintRpcAddFunctions()

	sReg.Conn.Start()
	defer sReg.Conn.Close()

	for {
		select {
		case <-time.After(6 * time.Second):
			go func() {
				args := []reflect.Value {reflect.ValueOf(time.Now().Unix())}
				rpcClient := register.NewRpcClient(register.RpcClient{NameSpace: "test"})
				result := rpcClient.Call("ping", args)
				fmt.Println("rpc返回", result)
			}()
		}
	}
}