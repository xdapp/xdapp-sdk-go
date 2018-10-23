package service

import (
	"fmt"
	"crypto/sha1"
	"encoding/hex"
	"time"
	"strconv"
)

type Sys struct {
	Register IRegister
}

func (service *Sys) getApp() string {
	return service.Register.GetApp()
}
func (service *Sys) getName() string {
	return service.Register.GetName()
}
func (service *Sys) getVersion() string {
	return service.Register.GetVersion()
}
func (service *Sys) getKey() string {
	return service.Register.GetKey()
}

/**
注册服务，在连接到 console 微服务系统后，会收到一个 sys_reg() 的rpc回调
*/
func (service *Sys) Reg(time int64, rand string, hash string) map[string]interface{} {

	// 验证hash
	if Sha1(fmt.Sprintf("%s.%s.%s", IntToStr(time), rand, "xdapp.com")) != hash {
		return nil
	}

	// 超时
	if Time() - time > 180 {
		return nil
	}

	app     := service.getApp()
	key     := service.getKey()
	name    := service.getName()
	version := service.getVersion()

	time = Time()
	hash = getHash(app, name, IntToStr(time), rand, key)
	return map[string]interface{}{"app": app, "name": name, "time": time, "rand": rand, "version": version,"hash": hash}
}

/**
获取菜单列表
*/
func (service *Sys) Menu() {

}

/**
注册失败
*/
func (service *Sys) RegErr(msg string, data interface{}) {
	service.Register.SetRegSuccess(false)
	service.Register.Error("注册失败", msg, data)
}

/**
注册成功回调
*/
func (service *Sys) RegOk(data map[string]map[string]string, time int, rand string, hash string) {

	app  := service.getApp()
	key  := service.getKey()
	name := service.getName()

	if getHash(app, name, IntToStr(time), rand, key) != hash {
		service.Register.SetRegSuccess(false)
		service.Register.CloseClient()
		return
	}

	// 注册成功
	service.Register.SetRegSuccess(true)
	service.Register.SetServiceData(data)

	service.Register.Debug("RPC服务注册成功，服务名:" + app + "-> " + name)

	// 同步页面
	service.Register.ConsolePageSync()
}

/**
测试接口
*/
func (service *Sys) Test(str string) {
	fmt.Println(str)
}

/**
获取rpc方法列表
*/
func (service *Sys) GetFunctions() []string {
	return service.Register.GetFunctions()
}

/**
hash 值
*/
func getHash(app string, name string, time string, rand string, key string) string {
	return Sha1(fmt.Sprintf("%s.%s.%s.%s.%s.xdapp.com", app, name, time, rand, key))
}

/**
获取sha1加密
*/
func Sha1(str string) string {
	getHash := sha1.New()
	getHash.Write([]byte(str))
	r := getHash.Sum(nil)
	return hex.EncodeToString(r[:])
}

/**
当前时间
*/
func Time() int64 {
	return time.Now().Unix()
}

func IntToStr(data interface{}) string {

	switch value := data.(type) {
	case int:
		return strconv.Itoa(value) // int to str
	case int64:
		return strconv.FormatInt(value, 10) // int64 转str
	default:
		return ""
	}
}