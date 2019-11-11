package service

import (
	"fmt"
	"reflect"
	"strconv"
	. "time"

	"crypto/sha1"
	"encoding/hex"
)

type IRegister interface {
	GetApp() string
	GetKey() string
	GetName() string
	GetVersion() string
	GetFunctions() []string
	SetRegSuccess(status bool)
	SetServiceData(data interface{}) error
	CloseClient()
	RpcCall(name string, args []reflect.Value, namespace string, cfg map[string]uint32) interface{}
	ILogger
}

type ILogger interface {
	Info(arg0 interface{}, args ...interface{})
	Debug(arg0 interface{}, args ...interface{})
	Warn(arg0 interface{}, args ...interface{})
	Error(arg0 interface{}, args ...interface{})
}

type SysService struct {
	Register IRegister
}

// 注册服务，在连接到 console 微服务系统后，会收到一个 sys_reg() 的rpc回调
func (service *SysService) Reg(time int64, rand string, hash string) map[string]interface{} {

	curHash := Sha1(fmt.Sprintf("%s.%s.%s", IntToStr(time), rand, "xdapp.com")); if curHash != hash {
		return map[string]interface{}{"status": false}
	}
	if Now().Unix() - time > 180 {
		return map[string]interface{}{"status": false}
	}

	app := service.Register.GetApp()
	key := service.Register.GetKey()
	name := service.Register.GetName()
	version := service.Register.GetVersion()

	time = Now().Unix()
	hash = getHash(app, name, IntToStr(time), rand, key)
	return map[string]interface{}{"status": true, "app": app, "name": name, "time": time, "rand": rand, "version": version,"hash": hash}
}

// 获取菜单列表
func (service *SysService) Menu() {

}

func (service *SysService) RegErr(msg string, data interface{}) {
	service.Register.SetRegSuccess(false)
	service.Register.Error("注册失败", msg, data)
}

func (service *SysService) RegOk(data interface{}, time int, rand string, hash string) {

	app := service.Register.GetApp()
	key := service.Register.GetKey()
	name := service.Register.GetName()

	if getHash(app, name, IntToStr(time), rand, key) != hash {
		service.Register.SetRegSuccess(false)
		service.Register.CloseClient()
		return
	}

	// 注册成功
	service.Register.SetRegSuccess(true)

	err := service.Register.SetServiceData(data)
	if err != nil {
		service.Register.Warn(err.Error())
	} else {
		service.Register.Debug("RPC服务注册成功，服务名:" + app + "-> " + name)
	}
}

// 测试接口
func (service *SysService) Test(str string) {
	fmt.Println(str)
}

func (service *SysService) Ping() bool {
	return true
}

// 获取rpc方法列表
func (service *SysService) GetFunctions() []string {
	return service.Register.GetFunctions()
}

// hash 值
func getHash(app string, name string, time string, rand string, key string) string {
	return Sha1(fmt.Sprintf("%s.%s.%s.%s.%s.xdapp.com", app, name, time, rand, key))
}

// 获取sha1加密
func Sha1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	r := h.Sum(nil)
	return hex.EncodeToString(r[:])
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
