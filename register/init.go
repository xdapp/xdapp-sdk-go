package register

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/alecthomas/log4go"
	"github.com/hprose/hprose-golang/rpc"
	"github.com/leesper/tao"
	"github.com/xdapp/xdapp-sdk-go/pkg/types"
)

type Config struct {
	App              string // 游戏简称
	Name             string // 游戏名字
	Key              string // 服务器秘钥
	Version          int    // 服务器版本
	IsDebug          bool   // 是否debug模式
	LogName          string // log文件名
	PackageMaxLength int    // tcp最大长度
}

type register struct {
	cfg         *Config
	Conn        *tao.ClientConn             // tcp客户端连接
	Logger      *log4go.Logger              // log 日志
	RegSuccess  bool                        // 注册成功标志
	ServiceData map[interface{}]interface{} // console 注册成功返回的页面服务器信息
}

var (
	config *Config
	Conn   *tao.ClientConn // tcp客户端连接
	Logger *log4go.Logger  // log 日志
)

func New(cfg *Config) (*register, error) {
	config = cfg

	if cfg.App == "" {
		return nil, types.ErrRequireApp
	}
	if cfg.Name == "" {
		return nil, types.ErrRequireServiceName
	}
	if cfg.Key == "" {
		return nil, types.ErrRequireServiceKey
	}

	if cfg.Version == 0 {
		cfg.Version = types.RPCVersion
	}
	if cfg.PackageMaxLength == 0 {
		cfg.PackageMaxLength = types.PackageMaxLength
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
	s := types.ProductionServer
	reg.ConnectTo(s.Host, s.Port, s.Ssl)
}

func (reg *register) ConnectToGlobal() {
	s := types.GlobalServer
	reg.ConnectTo(s.Host, s.Port, s.Ssl)
}

func (reg *register) ConnectToDev() {
	s := types.DevServer
	reg.ConnectTo(s.Host, s.Port, s.Ssl)
}

func (reg *register) ConnectTo(host string, port int, ssl bool) {
	Conn = NewClient(host, port, ssl)
	reg.Conn = Conn
	reg.Conn.Start()
	defer reg.Conn.Close()

	reg.Logger.Info(fmt.Sprintf("已增加的rpc列表, %v", GetHproseAddedFunc()))
	hproseService.AddMissingMethod(func(name string, args []reflect.Value, context rpc.Context) (result []reflect.Value, err error) {
		return nil, errors.New("The method '" + name + "' is not implemented.")
	})

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
func (reg *register) RpcCall(name string, args []reflect.Value, namespace string, cfg map[string]uint32) (interface{}, error) {
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
		logName = types.LogFileName
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
