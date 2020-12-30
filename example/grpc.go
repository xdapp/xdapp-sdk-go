package main

import (
	"github.com/xdapp/xdapp-sdk-go/pkg/middleware"
	"github.com/xdapp/xdapp-sdk-go/pkg/register"
	"google.golang.org/grpc"
)

func main() {
	reg, err := register.New(&register.Config{
		App:     "test", // 请修改对应的App缩写
		Name:    "test", // 请填入服务名，若协议Package为xdapp.api.v1则填入xdapp即可
		Key:     "test", // 从服务管理中添加服务后获取
		Debug: false,
	})

	if err != nil {
		panic(err)
	}

	// grpc service IP地址
	address := "localhost:7777"
	// grpc协议描述文件，参考：https://github.com/fullstorydev/grpcurl#protoset-files
	descriptor := []string{"./example/grpc/test.protoset"}
	proxy, err := middleware.NewGRPCProxyMiddleware(address, descriptor, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	reg.AddBeforeFilterHandler(proxy.Handler)

	reg.ConnectToDev()
}
