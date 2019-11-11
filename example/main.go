package main

import (
	"xdapp-sdk-go/register"
	"xdapp-sdk-go/service"
)

// 测试注册服务
func main() {
	reg, err := register.New(&register.Config{
		App: "demo",
		Name: "name",
		Key: "aaaaaaaaaa",
		IsDebug: false,
	})
	if err != nil {
		panic(err.Error())
	}

	// 注册系统rpc方法 （必加, 其中 sys 为服务前缀，请求方法为 sys_xxx）
	sysService := &service.SysService{Register: reg}
	register.AddInstanceMethods(sysService, "sys")

	//注册测试rpc方法 （其中 test 为服务前缀，请求方法为 test_xxx）
	testService := &service.TestService{Name: "test"}
	register.AddInstanceMethods(testService, "test")

	// 注册一个方法 访问 hello
	register.AddFunction("hello", func() string {return "hello world"}, "")

	register.Logger.Info("已增加的rpc列表", register.GetHproseAddedFunc())

	reg.ConnectTo("127.0.0.1", 8900, true)

	// 连接到外网测试服务器
	// reg.ConnectToDev()

	// 连接到生产环境(国内项目)
	//reg.ConnectToProduce()

	// 连接到生产环境(海外项目)
	// reg.ConnectToGlobal()
}