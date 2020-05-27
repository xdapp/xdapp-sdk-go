package main

import (
	"github.com/xdapp/xdapp-sdk-go/pkg/middleware"
	"github.com/xdapp/xdapp-sdk-go/register"
	"github.com/xdapp/xdapp-sdk-go/service"
	"google.golang.org/grpc"
)

func main() {
	proxy, err := middleware.NewGRPCProxyMiddleware("localhost:20002",
		[]string{"./example/service.protoset"}, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	reg, err := register.New(&register.Config{
		App:     "test", // 请修改配置参数
		Name:    "test",
		Key:     "test",
		IsDebug: false,
	})
	if err != nil {
		panic(err)
	}

	register.AddSysFunction(&service.SysService{Register: reg})
	register.AddBeforeFilterHandler(proxy.Handler)

	reg.ConnectToDev()
}
