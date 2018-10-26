package register

import (
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/alecthomas/log4go"
	"github.com/leesper/tao"
)

// 配置文件结构
type Config struct {
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
	Config
	Conn        *tao.ClientConn				   // tcp客户端连接
	Logger      *log4go.Logger                 // log 日志
	RegSuccess  bool                           // 注册成功标志
	ServiceData map[string]map[string]string   // console 注册成功返回的页面服务器信息
	ConsolePath []string 					   // console前端文件目录
}

// 可配置参数
type RegConfig struct {
	IsDebug             bool     `是否debug模式`
	LogName             string   `log文件名`
	ConfigPath          string   `配置文件路径`
	ConsolePath         []string `console前端文件目录`
	TcpConfig
}

// tcp配置
type TcpConfig struct {
	packageLengthOffset int      `tcp包长度位位置`
	packageBodyOffset   int      `tcp消息体位置`
	packageMaxLength    int      `tcp最大长度`
	tcpVersion          int      `tcp协议版本`
	host                string   `tcp请求地址`
}

const (
	DEFAULT_VER      = "1"
	DEFAULT_HOST     = "www.xdapp.com:8900"
	DEFAULT_SSL      = true
	DEFAULT_APP      = "test"
	DEFAULT_NAME     = "console"
	DEFAULT_KEY      = ""
	DEFAULT_LOG_NAME = "test.log"

	// 标识   | 版本    | 长度    | 头信息       | 自定义上下文  |  正文
	// ------|--------|---------|------------|-------------|-------------
	// Flag  | Ver    | Length  | Header     | Context     | Body
	// 1     | 1      | 4       | 17         | 默认0不定    | 不定
	// C     | C      | N       |            |             |
	// length 包括 Header + Context + Body 的长度

	DEFAULT_PACKAGE_LENGTH_OFFSET = 2        // 包长度开始位置
	DEFAULT_PACKAGE_BODY_OFFSET   = 6        // 包主体开始位置
	DEFAULT_PACKAGE_MAX_LENGTH    = 0x21000  // 最大的长度
)

var (
	Conn   *tao.ClientConn // tcp客户端连接
	Logger *log4go.Logger  // log 日志
	rpcCallRespMap  = make (map[string]chan interface{})
)

/**
创建
*/
func New(rfg RegConfig) (*SRegister, error) {

	Logger = NewLog4go(rfg.IsDebug, rfg.LogName)

	// tcp 配置
	if rfg.packageLengthOffset == 0 {
		rfg.packageLengthOffset = DEFAULT_PACKAGE_LENGTH_OFFSET
	}
	if rfg.packageBodyOffset == 0 {
		rfg.packageBodyOffset = DEFAULT_PACKAGE_BODY_OFFSET
	}
	if rfg.packageMaxLength == 0 {
		rfg.packageMaxLength = DEFAULT_PACKAGE_MAX_LENGTH
	}

	// console 前端目录
	if rfg.ConsolePath == nil {
		rfg.ConsolePath = append([]string{}, defaultBaseDir() + "/console/")
	}
	rfg.ConsolePath = checkExist(rfg.ConsolePath)

	if rfg.ConfigPath == "" {
		rfg.ConfigPath = defaultBaseDir() + "/config.yml"
	}

	conf, err := LoadConfig(rfg.ConfigPath)
	if err != nil {
		return nil, err
	}

	if rfg.host == "" {
		rfg.host = conf.Console.Host
	}

	return &SRegister{
		conf,
		NewClient(rfg),
		Logger,
		false,
		make(map[string]map[string]string),
		rfg.ConsolePath,
	}, nil
}

/**
设置配置
*/
func LoadConfig(filePath string) (Config, error) {

	if !PathExist(filePath) {
		return Config{}, fmt.Errorf("配置文件:%s 不存在", filePath)
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("读取配置文件错误:%s", err.Error())
	}

	config := Config{
		console{
			DEFAULT_HOST,
			DEFAULT_SSL,
			DEFAULT_APP,
			DEFAULT_NAME,
			DEFAULT_KEY,
		},
		DEFAULT_VER,
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("解析配置文件错误:%s", err.Error())
	}
	return config, nil
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