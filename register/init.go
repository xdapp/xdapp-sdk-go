package register

import (
	"fmt"
	"log"
	"strings"
	"io/ioutil"
	"path/filepath"
	"gopkg.in/yaml.v2"
	"github.com/alecthomas/log4go"
)
/**
配置
*/
type configuration struct {
	Console console
	Version string
}
type console struct {
	Host string // 服务器域名和端口
	SSl  bool   // 是否SSL连接
	App  string
	Name string
	Key  string
}

type SRegister struct {
	configuration
	Logger      *log4go.Logger                 // log 日志
	Client      *Client                        // 创建的tcp客户端对象
	RegSuccess  bool                           // 注册成功标志
	ServiceData (map[string]map[string]string) // console 注册成功返回的页面服务器信息
}

/**
可配置参数
*/
type RegConfig struct {
	IsDebug             bool     `是否debug模式`
	LogName             string   `log文件名`
	ConfigPath          string   `配置文件路径`
	ConsolePath         []string `console前端文件目录`
	packageLengthOffset int      `tcp包长度位位置`
	packageBodyOffset   int      `tcp消息体位置`
	packageMaxLength    int      `tcp最大长度`
}

const (
	defaultVersion             = "1"
	defaultHost                = "www.xdapp.com:8900"
	defaultSSl                 = true
	defaultApp                 = "test"
	defaultName                = "console"
	defaultKey                 = ""
	defaultLogName             = "test.log"

	// 标识   | 版本    | 长度    | 头信息       | 自定义上下文  |  正文
	// ------|--------|---------|------------|-------------|-------------
	// Flag  | Ver    | Length  | Header     | Context     | Body
	// 1     | 1      | 4       | 17         | 默认0不定    | 不定
	// C     | C      | N       |            |             |
	// length 包括 Header + Context + Body 的长度

	defaultPackageLengthOffset = 2        // 包长度开始位置
	defaultPackageBodyOffset   = 6        // 包主体开始位置
	defaultPackageMaxLength    = 0x21000  // 最大的长度
)

var Logger *log4go.Logger // log 日志

var rpcCallChan = make(chan interface{}, 10)

var rpcCallChan1 map[string]chan interface{}

/**
创建
*/
func New(rfg RegConfig) (*SRegister, error) {

	if rfg.LogName == "" {
		rfg.LogName = defaultLogName
	}
	Logger = NewLog4go(rfg.IsDebug, rfg.LogName)

	// tcp 配置
	if rfg.packageLengthOffset == 0 {
		rfg.packageLengthOffset = defaultPackageLengthOffset
	}
	if rfg.packageBodyOffset == 0 {
		rfg.packageBodyOffset = defaultPackageBodyOffset
	}
	if rfg.packageMaxLength == 0 {
		rfg.packageMaxLength = defaultPackageMaxLength
	}

	// console 前端目录
	if rfg.ConsolePath == nil {
		rfg.ConsolePath = defaultConsolePath()
	}
	setConsolePath(rfg.ConsolePath)

	if rfg.ConfigPath == "" {
		rfg.ConfigPath = defaultConfigPath()
	}

	conf, err := LoadConfig(rfg.ConfigPath)
	if err != nil {
		return nil, err
	}

	client := NewClient(conf.Console.Host,
		tcpConfig{
		rfg.packageLengthOffset,
		rfg.packageBodyOffset,
		rfg.packageMaxLength,
	})

	return &SRegister{
		conf,
		Logger,
		client,
		false,
		make(map[string]map[string]string,
		)}, nil
}

/**
设置配置
*/
func LoadConfig(filePath string) (configuration, error) {

	if !PathExist(filePath) {
		return configuration{}, fmt.Errorf("配置文件:%s 不存在", filePath)
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return configuration{}, fmt.Errorf("读取配置文件错误:%s", err.Error())
	}

	config := configuration{
		console{defaultHost, defaultSSl, defaultApp, defaultName, defaultKey},
		defaultVersion}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return configuration{}, fmt.Errorf("解析配置文件错误:%s", err.Error())
	}
	return config, nil
}

/**
默认基础目录
*/
func defaultBaseDir() string {
	dir, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func defaultConfigPath() string {
	return defaultBaseDir() + "/config.yml"
}

/**
默认前端目录
*/
func defaultConsolePath() []string {
	return append([]string{}, defaultBaseDir() + "/console/")
}


func (reg *SRegister) GetApp() string {
	return reg.Console.App
}

func (reg *SRegister) GetName() string {
	return reg.Console.Name
}
func (reg *SRegister) GetVersion() string {
	return reg.Version
}
func (reg *SRegister) GetKey() string {
	return reg.Console.Key
}

func (reg *SRegister) SetRegSuccess(status bool) {
	reg.RegSuccess = status
}

func (reg *SRegister) SetServiceData(data map[string]map[string]string) {
	reg.ServiceData = data
}

func (reg *SRegister) GetFunctions() []string {
	return RpcService.MethodNames
}

func (reg *SRegister) CloseClient() {
	reg.Client.Close(reg.RegSuccess)
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

/**
获取key
*/
func (reg *SRegister) getKey() string {
	return reg.ServiceData["pageServer"]["key"]
}

/**
获取host
*/
func (reg *SRegister) getHost() string {
	return reg.ServiceData["pageServer"]["host"]
}

/**
log4go对象设置
*/
func NewLog4go(isDebug bool, logName string) *log4go.Logger {

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