package register

import (
	"errors"
	"github.com/alecthomas/log4go"
	"github.com/leesper/tao"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

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

type TcpConfig struct {}

type register struct {
	cfg         *Config
	Conn        *tao.ClientConn               // tcp客户端连接
	Logger      *log4go.Logger                // log 日志
	RegSuccess  bool                          // 注册成功标志
	ServiceData map[interface{}]interface{}   // console 注册成功返回的页面服务器信息
}

const (
	DefaultVer              = 1
	DefaultApp              = "test"
	DefaultName             = "console"
	DefaultKey              = ""
	DefaultLogName          = "test.log"
	DefaultPackageMaxLength = 0x21000 // 最大的长度

	FlagSysMsg     = 1 // 来自系统调用的消息请求
	FlagResultMode = 2 // 请求返回模式，表明这是一个RPC结果返回
	FlagFinish     = 4 // 是否消息完成，用在消息返回模式里，表明RPC返回内容结束
	FlagTransport  = 8 // 转发浏览器RPC请求，表明这是一个来自浏览器的请求

	PrefixLength    = 6                           // Flag 1字节、 Ver 1字节、 Length 4字节
	HeaderLength    = 17                          // 默认消息头长度, 不包括 PrefixLength
	ContextOffset   = PrefixLength + HeaderLength // 自定义上下文内容所在位置，   23
	SendChunkLength = 0x200000                    // 单次发送的包大小
)

var (
	config *Config
	Conn   *tao.ClientConn // tcp客户端连接
	Logger *log4go.Logger  // log 日志

	ProductionServer = map[string]interface{}{
		"host": "service-prod.xdapp.com", "port": 8900, "ssl": true}

	DevServer = map[string]interface{}{
		"host": "dev.xdapp.com", "port": 8100, "ssl": true}

	GlobalServer = map[string]interface{}{
		"host": "service-gcp.xdapp.com", "port": 8900, "ssl": true}
)

func New(cfg *Config) (*register, error) {

	config = cfg

	if cfg.App == ""  {
		cfg.App = DefaultApp
	}
	if cfg.Name == ""  {
		cfg.Name = DefaultName
	}
	if cfg.Key == ""  {

		cfg.Key = DefaultKey
	}
	if cfg.Version == 0  {
		cfg.Version = DefaultVer
	}
	if cfg.PackageMaxLength == 0 {
		cfg.PackageMaxLength = DefaultPackageMaxLength
	}
	Logger = NewLog4go(cfg.IsDebug, cfg.LogName)

	return &register{
		cfg:         cfg,
		Logger:      Logger,
		RegSuccess:  false,
		ServiceData: nil,
	}, nil
}

func (reg *register) ConnectToProduce() {
	host := ProductionServer["host"].(string)
	port := ProductionServer["port"].(int)
	ssl := ProductionServer["ssl"].(bool)
	reg.ConnectTo(host, port, ssl)
}

func (reg *register) ConnectToGlobal() {
	host := GlobalServer["host"].(string)
	port := GlobalServer["port"].(int)
	ssl := GlobalServer["ssl"].(bool)
	reg.ConnectTo(host, port, ssl)
}

func (reg *register) ConnectToDev() {
	host := DevServer["host"].(string)
	port := DevServer["port"].(int)
	ssl := DevServer["ssl"].(bool)
	reg.ConnectTo(host, port, ssl)
}

func (reg *register) ConnectTo(host string, port int, ssl bool) {
	Conn = NewClient(host, port, ssl)
	reg.Conn = Conn
	reg.Conn.Start()
	defer reg.Conn.Close()

	notifier := make(chan os.Signal, 1)
	signal.Notify(notifier, syscall.SIGINT, syscall.SIGTERM)
	<-notifier
	os.Exit(0)
}

func (reg *register) getKey() string {
	pageSvr := reg.ServiceData["pageServer"].(map[string]string)
	return pageSvr["key"]
}

func (reg *register) getHost() string {
	pageSvr := reg.ServiceData["pageServer"].(map[string]string)
	return pageSvr["host"]
}

func (reg *register) GetApp() string {
	return reg.cfg.App
}

func (reg *register) GetName() string {
	return reg.cfg.Name
}
func (reg *register) GetVersion() string {
	return IntToStr(reg.cfg.Version)
}
func (reg *register) GetKey() string {
	return reg.cfg.Key
}

func (reg *register) SetRegSuccess(isReg bool) {
	reg.RegSuccess = isReg
}

func (reg *register) SetServiceData(data interface{}) error {
	svrData, ok := data.(map[interface{}]interface{})
	if !ok {
		return errors.New("regOK serviceData is illegal")
	}
	reg.ServiceData = svrData
	return nil
}

func (reg *register) GetFunctions() []string {
	return GetHproseAddedFunc()
}

func (reg *register) CloseClient() {
	reg.Conn.Close()
}

func (reg *register) Info(arg0 interface{}, args ...interface{}) {
	reg.Logger.Info(arg0, args...)
}

func (reg *register) Debug(arg0 interface{}, args ...interface{}) {
	reg.Logger.Debug(arg0, args...)
}

func (reg *register) Warn(arg0 interface{}, args ...interface{}) {
	reg.Logger.Warn(arg0, args...)
}

func (reg *register) Error(arg0 interface{}, args ...interface{}) {
	reg.Logger.Error(arg0, args...)
}

// 调取rpc服务
func (reg *register) RpcCall(name string, args []reflect.Value, namespace string, cfg map[string]uint32) interface{} {
	var serviceId uint32
	if _, ok := cfg["serviceId"]; ok {
		serviceId = cfg["serviceId"]
	}
	var adminId uint32
	if _, ok := cfg["adminId"]; ok {
		adminId = cfg["adminId"]
	}

	rpc := NewRpcClient(reg.Conn, serviceId, adminId, 0, namespace)
	return rpc.Call(name, args)
}

func NewLog4go(isDebug bool, logName string) *log4go.Logger {
	if logName == "" {
		logName = DefaultLogName
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