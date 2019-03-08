安装
----------
```
go get hub000.xindong.com/core-system/server-register-go
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


执行流程
----------
- 设置参数
- 通过hprose注册rpc方法 
- tcp连接到console服务 
- 执行reg检验参数
- 成注册登记 回调reg_ok
- 检查console目录的前端文件 + 同步更新
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
	"time"
	"reflect"
	"fmt"
	"server-register-go/register"
	"server-register-go/service"
	"github.com/hprose/hprose-golang/rpc"
)

func main() {
	sReg, err := register.New(register.Config{
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
	hproseService := register.HproseService
	hproseService.AddInstanceMethods(&service.Sys{sReg}, rpc.Options{NameSpace: "sys"})
	hproseService.AddInstanceMethods(&service.Test{"test service"}, rpc.Options{NameSpace: "test"})

	hproseService.AddFunction("hello", func() string {
		return "hello world"
	})

	sReg.Connect()
	defer sReg.Conn.Close()

	for {
		select {
		case <-time.After(5 * time.Second):
			go func() {
				args := []reflect.Value {reflect.ValueOf(time.Now().Unix())}
				result := sReg.RpcCall("ping", args, "test", map[string]uint32{})
				fmt.Println("rpc返回", result)
			}()
		}
	}
}
```