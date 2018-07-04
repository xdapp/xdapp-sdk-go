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
    1、rpc注册service文件夹中服务 （区分sys系统服务 + 普通服务）
    2、连接到consoloe tcp服务 执行reg
    3、成注册登记 回调reg_ok
    3、同步更新console 目录下vue文件


Example
----------
```golang
package main

import (
	"hub000.xindong.com/core-system/server-register-go/register"
	"hub000.xindong.com/core-system/server-register-go/service"
)

/**
	 测试注册服务
 */

func main() {
	register.LoadConfig()

	register.SetDebug(true)

	myReg := register.NewRegister()
	register.LoadService("", service.NewService(myReg))
	register.LoadService("sys", service.NewSysService(myReg))

	myReg.CreateClient()
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
