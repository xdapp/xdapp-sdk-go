安装
----------
```
go get github.com/xdapp/xdapp-sdk-go
```

rpc服务
----------

[hprose](https://github.com/hprose/hprose-golang/)


包管理
----------
    安装： brew install dep
    初始化： dep init

功能
----------
- 接入后台console服务器
- rpc注册service文件夹中服务 （区分sys系统服务 + 普通服务）
- 支持转发GRPC协议

服务器列表
----------

* 国内生产环境 `service-prod.xdapp.com:8900`
* 海外生产环境 `service-gcp.xdapp.com:8900`
* 开发测试环境 `service-dev.xdapp.com:8100`
* 本地测试环境 `127.0.0.1:8062`，需自己启动本地开发工具，see https://github.com/xdapp/xdapp-local-dev

> 除本地开发环境不是SSL的其它都是SSL的

执行流程
----------
- 设置参数
- 通过hprose注册rpc方法 
- tcp连接到console服务 
- 执行reg检验参数
- 成注册登记 回调reg_ok
- 调取rpc服务 rpc.Call xxx

> 如请求console test_ping(time)方法

```go
go func() {
    args := []reflect.Value {reflect.ValueOf(time.Now().Unix())}
    rpcClient := register.NewRpcClient(register.RpcClient{NameSpace: "test"})
    result := rpcClient.Call("ping", args)
    
    //返回 [pong, xxx] 
    fmt.Println("rpc返回", result)
}()

```


关于 `context` 上下文对象
----------

在RPC请求时，如果需要获取到请求时的管理员ID等等参数，可以用此获取，如上面 `hello` 的例子，通过 `context := register.getCurrentContext()` 可获取到 `context`，包括：

参数         |   说明               | 获取
------------|---------------------|---------------------
requestId   | 请求的ID             | context.GetInterface('requestId')
appId       | 请求的应用ID           | context.GetInterface('appId')
serviceId   | 请求发起的服务ID，0表示XDApp系统请求，1表示来自浏览器的请求    | context.GetInterface('serviceId')
adminId     | 请求的管理员ID，0表示系统请求 context.GetInterface('adminId')  |  context.GetInterface('adminId')
userdata    | 默认 stdClass 对象，可以自行设置参数   | context.UserData()


Example
----------
```golang
package main

import (
    "github.com/xdapp/xdapp-sdk-go/register"
    "github.com/xdapp/xdapp-sdk-go/service"
)

// 测试注册服务
func main() {
    // 其中 demo 为项目名，gm 为服务名，aaaaaaaaaa 为密钥
    reg, err := register.New(&register.Config{
        App: "demo",
        Name: "gm",
        Key: "aaaaaaaaaa",
        IsDebug: false,
    })
    if err != nil {
        panic(err)
    }

    // 注册系统rpc方法 （必加 用于sdk与xdapp服务注册）
    register.AddSysFunction(
        &service.SysService{Register: reg})

    /**
     * 注册单个前端页面可访问的rpc方法 （内部会加上服务名gm作前缀）
     * (!!! 请注意，只有服务名相同的前缀rpc方法才会被页面前端调用到)
     * 等同于
     * register.AddFunction("gm_hello", func() string {return "hello world"})
     * 页面请求方法 hello
     */
    register.AddWebFunction("hello", func() string {return "hello world"})

    /**
     * 注册某个struct下所有对外的方法 （内部会加上服务名前缀gm）
     * namespace: test, 页面请求方法 test_xxx
     * namespace可传空,  页面请求方法 xxx
     */
    register.AddWebInstanceMethods(
        &service.TestService{Name: "test"}, "test")

    reg.ConnectTo("127.0.0.1", 8900, false)

    // 连接到外网测试服务器
    //reg.ConnectToDev()

    // 连接到生产环境(国内项目)
    //reg.ConnectToProduce()

    // 连接到生产环境(海外项目)
    // reg.ConnectToGlobal()
}
```

转发GRPC协议
----------

SDK支持转发GRPC，通过协议文件描述符反射的方式转发请求

* Console后台的服务名将为协议包根目录名称
* GRPC协议中的类型会按照谷歌定义的[JSON Mapping](https://developers.google.com/protocol-buffers/docs/proto3#json) 做转换
* 前端请求时注意带上完整包名请求，并且严格区分大小写

前端调用例子
```proto3
# pb协议
syntax = "proto3";

# 服务名：test
package test.api.v1;

service TextCheck {
    rpc HelloWorld(google.protobuf.Empty) returns(google.protobuf.Empty);
}
```
```js
// 前端全局调用
require('app').service.test.api.v1.TextCheck.HelloWorld()
```

```golang
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
```
