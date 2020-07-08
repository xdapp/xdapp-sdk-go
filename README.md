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
go mod tidy

功能
----------
- 接入后台console服务器
- rpc注册service服务 （区分sys系统服务 + 普通服务, v1.1.0系统服务已内部处理）
- 支持转发GRPC协议
- 支持logger自定义 默认使用zap
- 简化连接流程 直接connect读取server配置

日志库
----------
日志库默认使用zap (v1.1.0开始)，或者也可以使用自定义 Logger

参考默认logger如何自定义它 https://github.com/xdapp/xdapp-sdk-go/blob/master/pkg/register/logger.go

自定义logger, 只要实现以下 interface就可以替换默认日志

```golang
type logger interface {
    Debug(msg string)
    Info(msg string)
    Warn(msg string)
    Error(msg string)
}
```
reg.SetLogger(xxx)

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
- 完成校验，sdk可接受来自console server的请求

Example
----------

简化server连接版
```golang
package main
package main

import (
	"github.com/xdapp/xdapp-sdk-go/pkg/register"
	"github.com/xdapp/xdapp-sdk-go/pkg/types"
	"github.com/xdapp/xdapp-sdk-go/service"
)

// 测试注册服务
func main() {

	// 其中 demo 为项目名，gm 为服务名，aaaaaaaaaa 为密钥
	reg, err := register.New(&register.Config{
		// prod:正式环境|dev:测试环境|global:海外环境|自定义环境不传 设置Sever参数
		// Env: "prod",
		
        // 自定义测试环境参数
		Server: &types.Server{
			Host: "127.0.0.1",
			Port: 8900,
			Ssl: false,
		},

		App: "demo",
		Name: "gm",
		Key: "xx",
		LogOutputs: []string{"stderr", "debug"},
		Debug: true,
	})
	if err != nil {
		panic(err)
	}

	/**
	 * 注册单个前端页面可访问的rpc方法 （内部会加上服务名gm作前缀）
	 * (!!! 请注意，只有服务名相同的前缀rpc方法才会被页面前端调用到)
	 * 等同于
	 * register.AddFunction("gm_hello", func() string {return "hello world"})
	 * 页面请求方法 hello
	 */
	reg.AddWebFunction("hello", func() string {return "hello world"})

	/**
	 * 注册某个struct下所有对外的方法 （内部会加上服务名前缀gm）
	 * namespace: test, 页面请求方法 test_xxx
	 * namespace可传空,  页面请求方法 xxx
	 */
	reg.AddWebInstanceMethods(
		&service.TestService{Name: "test"}, "test")

	// 读取配置连接
	reg.Connect()
}

```

旧版本connectTo方式同样支持

```golang
package main

import (
    "github.com/xdapp/xdapp-sdk-go/pkg/register"
    "github.com/xdapp/xdapp-sdk-go/service"
)

// 测试注册服务
func main() {
    // 其中 demo 为项目名，gm 为服务名，aaaaaaaaaa 为密钥
    reg, err := register.New(&register.Config{
		App: "demo",
		Name: "gm",
		Key: "xx",
		LogOutputs: []string{"stderr", "debug"},
		Debug: true,
    })
    if err != nil {
        panic(err)
    }

    /**
     * 注册单个前端页面可访问的rpc方法 （内部会加上服务名gm作前缀）
     * (!!! 请注意，只有服务名相同的前缀rpc方法才会被页面前端调用到)
     * 等同于
     * register.AddFunction("gm_hello", func() string {return "hello world"})
     * 页面请求方法 hello
     */
    reg.AddWebFunction("hello", func() string {
        return "hello world"
    })

    /**
     * 注册某个struct下所有对外的方法 （内部会加上服务名前缀gm）
     * namespace: test, 页面请求方法 test_xxx
     * namespace可传空,  页面请求方法 xxx
     */
    reg.AddWebInstanceMethods(
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


```js
// 前端全局调用
require('app').service.hello()
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
	"github.com/xdapp/xdapp-sdk-go/pkg/register"
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
		Debug: false,
	})

	if err != nil {
		panic(err)
	}

	reg.AddBeforeFilterHandler(proxy.Handler)

	reg.ConnectToDev()
}
```

Changelog
----------

### v1.1.0
1. change register path;<br/>
调整register目录结构, 精简优化代码
2. change constant error;<br/>
调整 constant常量、error
3. feature default logger use zap;<br/>
默认日志库使用zap, 增加log
4. feature simplify AddSysFunction;<br/>
简化注册流程 系统方法 AddSysFunction 内部处理
5. feature simplify connect;<br/>
简化连接流程 直接connect读取server配置
6. feature coroutine handle rpc request;<br/>
RPC请求非阻塞 协程处理

### v1.0.5
1. feature add GRPC proxy;<br/>
转发GRPC协议