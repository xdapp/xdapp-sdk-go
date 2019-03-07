package main

import (
	"time"
	"server-register-go/register"
	"server-register-go/service"
	"github.com/hprose/hprose-golang/rpc"
	"reflect"
	"fmt"
)

/**
测试注册服务
*/
func main() {
	sReg, err := register.New(register.Config{
		App: "demo",
		Name: "name",
		SSl: false,
		Key: "aaaaaaaaaa",
		Host: "172.26.128.162:8900",
		IsDebug: false,
	})
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

	sReg.Connect()
	defer sReg.Conn.Close()

	for {
		select {
		case <-time.After(5 * time.Second):
			go func() {
				rpcClient := register.NewRpcClient(register.RpcClient{NameSpace: "test"})
				result := rpcClient.Call("ping",
					[]reflect.Value {reflect.ValueOf(time.Now().Unix())})
				fmt.Println("rpc返回", result)
			}()
		}
	}
}