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
    //reg.ConnectToProduce()

    // 连接到生产环境(国内项目)
    //reg.ConnectToProduce()

    // 连接到生产环境(海外项目)
    // reg.ConnectToGlobal()
}
```