package register

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/alecthomas/log4go"
	"fmt"
	"path/filepath"
	"log"
	"strings"
)

/**
	配置
 */
type Config struct {
	Console console
	Env string
}

type console struct {
	Host 	string // 服务器域名和端口
	SSl 	bool   // 是否SSL连接
	App  	string
	Name 	string
	Key  	string
}

const (
	defaultHostIsSSl   = true
	defaultServiceApp  = "test"
	defaultServiceName = "console"
	defaultHost        = "www.xdapp.com:8900"
)

/**
	全局变量
 */
var (
	conf Config				// 配置
	myRpc *MyRpc			// rpc 服务
	MyLog *log4go.Logger	// log日志
)
var logFile = "test.log"

/**
	全局变量初始化
 */
func init() {
	conf  = getConf(configPath())			// 配置
	MyLog = NewLog4go(conf.Env, logFile)	// 获取log对象
	myRpc = NewMyRpc()						// rpc服务
}

type RegisterData struct {
	console
	MyClient    *Client
	RegSuccess  bool
	ServiceData (map[string]map[string]string)
}

/**
	工厂创建
 */
func NewRegister() *RegisterData {

	Console := conf.Console
	if Console.Host == "" {
		Console.Host = defaultHost
	}
	if Console.SSl == false {
		Console.SSl = defaultHostIsSSl
	}
	if Console.App == "" {
		Console.App = defaultServiceApp
	}
	if Console.Name == "" {
		Console.Name = defaultServiceName
	}

	client := NewClient(Console.Host, *config)
	return &RegisterData{Console, client, false, make (map[string]map[string]string)}
}

/**
	获取key
 */
func (reg *RegisterData) getKey() string {
	return reg.ServiceData["pageServer"]["key"]
}

/**
	获取host
 */
func (reg *RegisterData) getHost() string {
	return reg.ServiceData["pageServer"]["host"]
}

/**
	配置文件
 */
func configPath() string {
	return GetBaseDir()  + "/config.yml"
}

/**
	获取配置
 */
func getConf(path string) Config {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("yamlFile.Get err   #%v ", err)
	}

	conf := Config{}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	return conf
}

/**
	获取一个类型的路径
 */
func GetPath(pathType string) []string {

	var path []string
	baseDir := GetBaseDir()

	path = append(path, baseDir + "/" + pathType)							// 项目
	path = append(path, baseDir + "/vendor/xdapp/game/" + pathType)			// 项目库
	path = append(path, baseDir + "/vendor/xdapp/game-core/" + pathType)	// 核心核心库
	return path
}

/**
	获取当前目录
 */
func GetBaseDir() string {
	dir, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

/**
	log4go对象设置
 */
func NewLog4go(env string, logFile string) *log4go.Logger {
	log4 := make(log4go.Logger)

	if env == "dev" {	// 开发环境log
		log4.AddFilter("stdout", log4go.DEBUG, log4go.NewConsoleLogWriter())
		log4.AddFilter("file", log4go.INFO, log4go.NewFileLogWriter(logFile, false))
	} else {			// 线上环境log
		log4.AddFilter("stdout", log4go.DEBUG, log4go.NewConsoleLogWriter())
		log4.AddFilter("file", log4go.DEBUG, log4go.NewFileLogWriter(logFile, false))
	}
	return &log4
}

func Debug(arg0 interface{}, args ...interface{}) {
	MyLog.Debug(arg0, args ...)
}

func Info(arg0 interface{}, args ...interface{}) {
	MyLog.Info(arg0, args ...)
}

func Error(arg0 interface{}, args ...interface{}) {
	MyLog.Error(arg0, args ...)
}

func (reg *RegisterData) GetApp() string {
	return reg.App
}

func (reg *RegisterData) GetName() string {
	return reg.Name
}
func (reg *RegisterData) GetKey() string {
	return reg.Key
}

func (reg *RegisterData) SetRegSuccess(status bool) {
	reg.RegSuccess = status
}

func (reg *RegisterData) SetServiceData(data map[string]map[string]string) {
	reg.ServiceData = data
}

func (reg *RegisterData) CloseClient() {
	reg.MyClient.Close(reg.RegSuccess)
}