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


Example
----------
```golang
package main

import (
	"github.com/xdapp/xdapp-sdk-go/register"
	"github.com/xdapp/xdapp-sdk-go/service"
	"fmt"
	"time"
)

// 测试注册服务
func main() {
	reg, err := register.New(register.Config{
		App: "demo",
		Name: "name",
		SSl: false,
		Key: "123456",
		Host: "service-dev.xdapp.com:8100",     // 其它环境服务器见服务器列表
		IsDebug: false,
	})
	if err != nil {
		panic(err.Error())
	}

	// 加载rpc 方法
	register.AddInstanceMethods(&service.Sys{reg}, "sys")
	register.AddInstanceMethods(&service.Test{"test service"}, "test")

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
		// result := reg.RpcCall("ping", args, "test", map[string]uint32{})
		}
	}
}
```