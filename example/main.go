package main

import (
	"fmt"
	"time"
	"xdapp-sdk-go/register"
	"xdapp-sdk-go/service"
)

// 测试注册服务
func main() {
	reg, err := register.New(register.Config{
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
	register.AddInstanceMethods(&service.SysService{reg}, "sys")
	register.AddInstanceMethods(&service.TestService{"test service"}, "test")

	hproseService := register.HproseService
	hproseService.AddFunction("hello", func() string {
		return "hello world"
	})

	fmt.Println("已增加的rpc列表", register.GetHproseAddedFunc())

	reg.Conn.Start()
	defer reg.Conn.Close()

	for {
		select {
		case <-time.After(5 * time.Second):
			fmt.Println("测试")
		}
	}
}