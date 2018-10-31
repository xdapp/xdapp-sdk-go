安装
----------
```
go get hub000.xindong.com/core-system/server-register-go
```


包管理
----------
    安装： brew install dep
    初始化： dep init

功能
----------
- rpc注册service文件夹中服务 （区分sys系统服务 + 普通服务）
- tcp连接到console服务 执行reg检验参数
- 成注册登记 回调reg_ok
- 检查console目录的前端文件 + 同步更新
- 调取rpc服务 rpc.Call xxx
> 如请求console player_test方法

```go
now := time.Now().Unix()
args :=[]reflect.Value {reflect.ValueOf(now)}
	
rpc := NewRpcCall(RpcCall{
        nameSpace: "player",
})
result := rpc.Call("test", args)

```


Example
----------
```golang
package main

import (
	"server-register-go/register"
	"server-register-go/service"
	"time"
)

/**
测试注册服务
*/
func main() {
	sReg, err := register.New(register.RegConfig{IsDebug: true})
	if err != nil {
		panic(err.Error())
	}

	// 加载rpc 方法
	register.AddInstanceMethods(&service.Sys{sReg}, "sys")
	register.AddInstanceMethods(&service.Test{"test service"}, "test")

	// 增加单个方法
	register.AddFunction("hello", func() string {
		return "hello world"
	})
	register.PrintRpcAddFunctions()

	sReg.Conn.Start()
	defer sReg.Conn.Close()

	for {
		select {
		case <-time.After(6 * time.Second):
			go register.TestRpcCall()
		}
	}
}

```
配置文件：

config.yml
```golang
console:
    app: test
    name: name
    key: aaaaaaaaaa
    host: 127.0.0.1:8900
    ssl: false
```
