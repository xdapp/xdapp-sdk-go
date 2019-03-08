package register

import (
	"github.com/alecthomas/log4go"
	"github.com/leesper/tao"
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
	ConsolePath      []string // console前端文件目录
	PackageMaxLength int      // tcp最大长度
}

// tcp配置
type TcpConfig struct {
}

type SRegister struct {
	Conn        *tao.ClientConn              // tcp客户端连接
	Logger      *log4go.Logger               // log 日志
	RegSuccess  bool                         // 注册成功标志
	ServiceData map[string]map[string]string // console 注册成功返回的页面服务器信息
	ConsolePath []string                     // console前端文件目录
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
	Conn   *tao.ClientConn // tcp客户端连接
	Logger *log4go.Logger  // log 日志
	startChan chan bool
	config Config
)

/**
创建
*/
func New(rfg Config) (*SRegister, error) {

	if rfg.Host == "" {
		rfg.Host = DEFAULT_HOST
	}
	if rfg.SSl == false {
		rfg.SSl = DEFAULT_SSL
	}
	if rfg.App == ""  {
		rfg.App = DEFAULT_APP
	}
	if rfg.Name == ""  {
		rfg.Name = DEFAULT_NAME
	}
	if rfg.Key == ""  {
		rfg.Name = DEFAULT_KEY
	}
	if rfg.Version == 0  {
		rfg.Version = DEFAULT_VER
	}

	if rfg.PackageMaxLength == 0 {
		rfg.PackageMaxLength = DEFAULT_PACKAGE_MAX_LENGTH
	}

	// console 前端目录
	if rfg.ConsolePath == nil {
		rfg.ConsolePath = append([]string{}, defaultBaseDir() + "/console/")
	}
	rfg.ConsolePath = checkExist(rfg.ConsolePath)

	config = rfg
	startChan = make(chan bool)
	Logger = NewLog4go(rfg.IsDebug, rfg.LogName)

	return &SRegister{
		NewClient(rfg.Host),
		Logger,
		false,
		make(map[string]map[string]string),
		rfg.ConsolePath,
	}, nil
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