package service

import (
	"fmt"
	"strings"
	"runtime"
	"path/filepath"
	"crypto/sha1"
	"encoding/hex"
	"time"
	"strconv"
)

type IRegister interface {
	GetApp() string
	GetName() string
	GetKey() string
	SetRegSuccess(status bool)
	SetServiceData(data map[string]map[string]string)
	CloseClient()
	ConsolePageSync()
	ILogger
}

/**
	logger
 */
type ILogger interface {
	Info(arg0 interface{}, args ...interface{})
	Debug(arg0 interface{}, args ...interface{})
	Warn(arg0 interface{}, args ...interface{})
	Error(arg0 interface{}, args ...interface{})
}

type SysService struct {
	register IRegister
}

func NewSysService(register IRegister) *SysService {
	return &SysService{register: register}
}

func (service *SysService) getApp() string {
	return service.register.GetApp()
}

func (service *SysService) getName() string {
	return service.register.GetName()
}

func (service *SysService) getKey() string {
	return service.register.GetKey()
}

/**
	注册服务，在连接到 console 微服务系统后，会收到一个 sys_reg() 的rpc回调
  */
func (service *SysService) Reg(time int64, rand string, hash string) []interface{} {

	// 当前方法名
	fun := strings.ToLower(GetFuncName())

	// 验证hash
	if Sha1(fmt.Sprintf("%s.%s.%s", IntToStr(time), rand, "xdapp.com")) != hash {
		return []interface{} {fun, false};
	}

	// 超时
	if Time() - time > 180 {
		return []interface{} {fun, false};
	}

	app  := service.getApp()
	key  := service.getKey()
	name := service.getName()
	time  = Time()
	hash  = getHash(app, name, IntToStr(time), rand, key)

	return []interface{} {fun, map[string]interface{}{"app": app, "name": name , "time": time, "rand": rand, "hash": hash}}
}

/**
	获取菜单列表
 */
func (service *SysService) Menu() {

}

/**
	注册失败
 */
func (service *SysService) RegErr(msg string, data interface{}) {
	service.register.SetRegSuccess(false)
	service.register.Error("注册失败", msg, data)
}

/**
	注册成功回调
 */
func (service *SysService) RegOk(data map[string]map[string]string, time int, rand string, hash string) {

	app  := service.getApp()
	key  := service.getKey()
	name := service.getName()

	if getHash(app, name, IntToStr(time), rand, key) != hash {
		service.register.SetRegSuccess(false)
		service.register.CloseClient()
		return
	}

	// 注册成功
	service.register.SetRegSuccess(true)
	service.register.SetServiceData(data)

	service.register.Debug("RPC服务注册成功，服务名:" + app + "-> " + name)

	// 同步页面
	service.register.ConsolePageSync()
}

/**
	测试接口
 */
func (service *SysService) Test(str string) {
	fmt.Println(str)
}

/**
	获取函数名
 */
func GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	funcName = filepath.Ext(funcName)
	funcName = strings.TrimPrefix(funcName, ".")
	return funcName;
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

/**
	hash 值
 */
func getHash(app string, name string, time string, rand string, key string) string {
	return Sha1(fmt.Sprintf("%s.%s.%s.%s.%s.xdapp.com", app, name, time, rand, key))
}