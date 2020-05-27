package main

import (
	"github.com/xdapp/xdapp-sdk-go/pkg/middleware"
	"github.com/xdapp/xdapp-sdk-go/register"
	"github.com/xdapp/xdapp-sdk-go/service"
	"google.golang.org/grpc"
)

func main() {
	// grpc service IP地址
	address := "localhost:8080"
	// grpc协议描述文件，参考：https://github.com/fullstorydev/grpcurl#protoset-files
	descriptor := []string{"./example/service.protoset"}
	proxy, err := middleware.NewGRPCProxyMiddleware(address, descriptor, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	reg, err := register.New(&register.Config{
		App:     "test", // 请修改对应的App缩写
		Name:    "test", // 请填入服务名，若协议Package为xdapp.api.v1则填入xdapp即可
		Key:     "test", // 从服务管理中添加服务后获取
		IsDebug: false,
	})

	if err != nil {
		panic(err)
	}

	register.AddSysFunction(&service.SysService{Register: reg})
	register.AddBeforeFilterHandler(proxy.Handler)

	reg.ConnectToDev()
}
