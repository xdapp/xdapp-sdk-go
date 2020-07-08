package main

import (
	"github.com/xdapp/xdapp-sdk-go/pkg/register"
	"github.com/xdapp/xdapp-sdk-go/service"
)

// 测试注册服务
func main() {
	reg, err := register.New(&register.Config{
		Env: "dev",
		App: "demo",
		Name: "gm",
		Key: "123456",
		Debug: false,
	})
	if err != nil {
		panic(err)
	}

	/**
	 * 注册一个前端页面可访问的rpc方法 （其中 test 为服务前缀，请求方法为 hello）
	 * (请注意，只有服务名相同的前缀rpc方法才会被页面前端调用到)
	 * 等同于
	 * register.AddFunction("gm_hello", func() string {return "hello world"})
	 */
	reg.AddWebFunction("hello", func() string {return "hello world"})

	/**
	 * 注册一个前端页面可访问的rpc方法 （其中 test 为服务前缀，请求方法为 test_xxx）
	 */
	reg.AddWebInstanceMethods(
		&service.TestService{Name: "test"}, "test")

	reg.Connect()
}