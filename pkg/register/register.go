package register

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/hprose/hprose-golang/rpc"
	"github.com/leesper/tao"
	"github.com/xdapp/xdapp-sdk-go/pkg/types"
	"github.com/xdapp/xdapp-sdk-go/service"
	"go.uber.org/zap"
)

type Config struct {
	*types.Server									// 自定义服务参数
	Env				 string 						// 环境 dev、prod、global 不传则读Server配置

	App              string   						// 游戏简称
	Name             string   						// 游戏名字
	Key              string   						// 服务器秘钥
	Version          int       						// 服务器版本
	Debug            bool     `json:"debug"`  		// debug模式
	PackageMaxLength int      						// tcp最大长度 默认2M
	LogOutputs       []string `json:"log_outputs"`	// 输出格式 默认 []string{"stdout"} 需要落地到debug.log eg: []string{"stdout", "debug.log"}

	loggerMu      *sync.RWMutex
	logger        *zap.Logger
	loggerConfig  *zap.Config
}

type clientConn struct {
	*tao.ClientConn
}

type register struct {
	lg            logger
	cfg           *Config
	conn          *clientConn                 // tcp客户端连接
	RegSuccess    bool                        // 注册成功标志
	ServiceData   map[interface{}]interface{} // console 注册成功返回的页面服务器信息
	HproseService *rpc.TCPService             // hprose service
}

var (
	lg     logger // zap logger
	config *Config
)

func New(cfg *Config) (*register, error) {
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

	err := cfg.setServerConfig()
	if err != nil {
		return nil, err
	}

	if len(cfg.LogOutputs) == 0 {
		cfg.LogOutputs = []string{"stdout"}
	}

	// set zap logger
	err = cfg.setupLogging()
	if err != nil {
		return nil, err
	}

	lg = cfg.Logger()
	config = cfg

	return &register{
		cfg:           cfg,
		lg:            lg,
		RegSuccess:    false,
		ServiceData:   nil,
		HproseService: rpc.NewTCPService(),
	}, nil
}

// set server config
func (cfg *Config) setServerConfig() error {
	if cfg.Env == types.EnvironmentProd {
		cfg.Server = types.ProductionServer
	} else if cfg.Env == types.EnvironmentDev {
		cfg.Server = types.DevServer
	} else if cfg.Env == types.EnvironmentGlobal {
		cfg.Server = types.GlobalServer
	}
	return nil
}

// connect server config
func (reg *register) Connect() {
	s := reg.cfg
	if s.Server == nil {
		panic(types.ErrRequireServer)
	}
	if s.Host == "" {
		panic(types.ErrRequireHost)
	}
	if s.Port == 0 {
		panic(types.ErrRequirePort)
	}
	reg.ConnectTo(s.Host, s.Port, s.Ssl)
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
	// add sys function
	reg.AddSysFunction(
		&service.SysService{Register: reg})

	msg := fmt.Sprintf("[tcp]连接 host:%s port:%d", host, port)
	reg.lg.Info(msg)

	conn := reg.NewClient(host, port, ssl)
	reg.conn = &clientConn{conn}
	reg.conn.Start()
	defer reg.conn.Close()

	funcArr := reg.GetHproseAddedFunc()
	reg.lg.Info("已增加的rpc方法: " + strings.Join(funcArr, ","))

	reg.HproseService.AddMissingMethod(
		func(name string, args []reflect.Value, context rpc.Context) (result []reflect.Value, err error) {
			return nil, errors.New("The method '" + name + "' is not implemented.")
		})

	notifier := make(chan os.Signal, 1)
	signal.Notify(notifier, syscall.SIGINT, syscall.SIGTERM)
	<-notifier
	os.Exit(0)
}

func (reg *register) GetApp() string {
	return reg.cfg.App
}

func (reg *register) GetName() string {
	return reg.cfg.Name
}
func (reg *register) GetVersion() string {
	return strconv.Itoa(reg.cfg.Version)
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
	return reg.GetHproseAddedFunc()
}

func (reg *register) CloseClient() {
	reg.conn.Close()
}

func (reg *register) Info(msg string) {
	reg.lg.Info(msg)
}

func (reg *register) Debug(msg string) {
	reg.lg.Debug(msg)
}

func (reg *register) Warn(msg string) {
	reg.lg.Warn(msg)
}

func (reg *register) Error(msg string) {
	reg.lg.Error(msg)
}

// 调取rpc服务
func (reg *register) RpcCall(
	name string,
	args []reflect.Value,
	namespace string,
	cfg map[string]uint32) (interface{}, error) {

	var serviceId uint32
	if _, ok := cfg["serviceId"]; ok {
		serviceId = cfg["serviceId"]
	}
	var adminId uint32
	if _, ok := cfg["adminId"]; ok {
		adminId = cfg["adminId"]
	}

	rpc := NewRpcClient(reg.conn.ClientConn, serviceId, adminId, 0, namespace)
	return rpc.Call(name, args)
}

// SetLogger replace default logger
func (reg *register) SetLogger(log logger) {
	reg.lg = log
	lg = log
}

// get logger
func (reg *register) GetLogger() logger {
	return reg.lg
}
