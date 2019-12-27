package main

import (
	"github.com/xdapp/xdapp-sdk-go/register"
	"github.com/xdapp/xdapp-sdk-go/service"
)

// 测试注册服务
func main() {
	reg, err := register.New(&register.Config{
		App: "ro",
		Name: "gm",
		Key: "123456",
		IsDebug: false,
	})
	if err != nil {
		panic(err)
	}

	// 注册系统rpc方法 （必加, 其中 sys 为服务前缀，请求方法为 sys_xxx）
	register.AddSysFunction(
		&service.SysService{Register: reg})

	/**
	 * 注册一个前端页面可访问的rpc方法 （其中 test 为服务前缀，请求方法为 hello）
	 * (请注意，只有服务名相同的前缀rpc方法才会被页面前端调用到)
	 * 等同于
	 * register.AddFunction("gm_hello", func() string {return "hello world"})
	 */
	register.AddWebFunction("hello", func() string {return "hello world"})

	/**
	 * 注册一个前端页面可访问的rpc方法 （其中 test 为服务前缀，请求方法为 test_xxx）
	 */
	register.AddWebInstanceMethods(
		&service.TestService{Name: "test"}, "test")

	reg.ConnectTo("127.0.0.1", 8900, false)

	// 连接到外网测试服务器
	//reg.ConnectToProduce()

	// 连接到生产环境(国内项目)
	//reg.ConnectToProduce()

	// 连接到生产环境(海外项目)
	// reg.ConnectToGlobal()
}