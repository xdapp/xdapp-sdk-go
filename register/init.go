package register

import (
	"errors"
	"github.com/alecthomas/log4go"
	"github.com/leesper/tao"
	"reflect"
)

// 可配置参数
type Config struct {
	Host             string   // 服务器域名和端口
	SSl              bool     // 是否SSL连接
	App              string   // 游戏简称
	Name             string   // 游戏名字
	Key              string   // 服务器秘钥
	Version          int      // 服务器版本
	IsDebug          bool     // 是否debug模式
	LogName          string   // log文件名
	PackageMaxLength int      // tcp最大长度
}

// tcp配置
type TcpConfig struct {
}

type SRegister struct {
	Conn        *tao.ClientConn              // tcp客户端连接
	Logger      *log4go.Logger               // log 日志
	RegSuccess  bool                         // 注册成功标志
	ServiceData map[interface{}]interface{} 		 // console 注册成功返回的页面服务器信息
}

const (
	DEFAULT_VER                = 1
	DEFAULT_HOST               = "www.xdapp.com:8900"
	DEFAULT_SSL                = true
	DEFAULT_APP                = "test"
	DEFAULT_NAME               = "console"
	DEFAULT_KEY                = ""
	DEFAULT_LOG_NAME           = "test.log"

	DEFAULT_PACKAGE_MAX_LENGTH = 0x21000 // 最大的长度
)

var (
	config Config
	Conn   *tao.ClientConn // tcp客户端连接
	Logger *log4go.Logger  // log 日志
)

/**
创建
*/
func New(rfg Config) (*SRegister, error) {

	config = rfg
	if config.Host == "" {
		config.Host = DEFAULT_HOST
	}
	if config.SSl == false {
		config.SSl = DEFAULT_SSL
	}
	if config.App == ""  {
		config.App = DEFAULT_APP
	}
	if config.Name == ""  {
		config.Name = DEFAULT_NAME
	}
	if config.Key == ""  {
		config.Name = DEFAULT_KEY
	}
	if config.Version == 0  {
		config.Version = DEFAULT_VER
	}
	if config.PackageMaxLength == 0 {
		config.PackageMaxLength = DEFAULT_PACKAGE_MAX_LENGTH
	}

	Logger = NewLog4go(config.IsDebug, config.LogName)
	Conn = NewClient(config.Host)

	return &SRegister{Conn,Logger,false,nil,}, nil
}

func (reg *SRegister) getKey() string {

	pageSvr := reg.ServiceData["pageServer"].(map[string]string)
	return pageSvr["key"]
}

func (reg *SRegister) getHost() string {
	pageSvr := reg.ServiceData["pageServer"].(map[string]string)
	return pageSvr["host"]
}

func (reg *SRegister) GetApp() string {
	return config.App
}

func (reg *SRegister) GetName() string {
	return config.Name
}
func (reg *SRegister) GetVersion() string {
	return IntToStr(config.Version)
}
func (reg *SRegister) GetKey() string {
	return config.Key
}

func (reg *SRegister) SetRegSuccess(isReg bool) {
	reg.RegSuccess = isReg
}

func (reg *SRegister) SetServiceData(data interface{}) error {
	svrData, ok := data.(map[interface{}]interface{})
	if !ok {
		return errors.New("regOK serviceData is illegal")
	}
	reg.ServiceData = svrData
	return nil
}

func (reg *SRegister) GetFunctions() []string {
	return GetHproseAddedFunc()
}

func (reg *SRegister) CloseClient() {
	reg.Conn.Close()
}

func (reg *SRegister) Info(arg0 interface{}, args ...interface{}) {
	reg.Logger.Info(arg0, args...)
}

func (reg *SRegister) Debug(arg0 interface{}, args ...interface{}) {
	reg.Logger.Debug(arg0, args...)
}

func (reg *SRegister) Warn(arg0 interface{}, args ...interface{}) {
	reg.Logger.Warn(arg0, args...)
}

func (reg *SRegister) Error(arg0 interface{}, args ...interface{}) {
	reg.Logger.Error(arg0, args...)
}

// 调取rpc服务
func (reg *SRegister) RpcCall(name string, args []reflect.Value, namespace string, cfg map[string]uint32) interface{} {
	var serviceId uint32
	if _, ok := cfg["serviceId"]; ok {
		serviceId = cfg["serviceId"]
	}
	var adminId uint32
	if _, ok := cfg["adminId"]; ok {
		adminId = cfg["adminId"]
	}

	rpc := NewRpcClient(RpcClient{
		NameSpace: namespace,
		ServiceId: serviceId,
		AdminId: adminId,
	})

	return rpc.Call(name, args)
}

/**
log4go对象设置
*/
func NewLog4go(isDebug bool, logName string) *log4go.Logger {
	if logName == "" {
		logName = DEFAULT_LOG_NAME
	}

	log4 := make(log4go.Logger)
	cw := log4go.NewConsoleLogWriter()

	// 非debug模式
	if isDebug == false {
		cw.SetFormat("[%T %D] [%L] %M")
	}
	log4.AddFilter("stdout", log4go.DEBUG, cw)
	log4.AddFilter("file", log4go.ERROR, log4go.NewFileLogWriter(logName, false))

	return &log4
}