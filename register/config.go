package register

import (
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"log"
	"strings"
)

/**
	配置
 */
type configuration struct {
	Console console
}

/**
	console
 */
type console struct {
	Host 	string // 服务器域名和端口
	SSl 	bool   // 是否SSL连接
	App  	string
	Name 	string
	Key  	string
}

/**
	console 默认配置
 */
var defaultConsole = console{
	"www.xdapp.com:8900",true,"test","console","",}

var (
	baseDir  = defaultBaseDir()
	confPath = defaultConfPath()
	conf     = configuration{defaultConsole} // 默认配置
)

/**
	设置配置
 */
func LoadConfig() configuration {

	if !PathExist(confPath) {
		MyLog.Error("配置文件：" + confPath + "不存在！")
	}

	data, err := ioutil.ReadFile(confPath)
	if err != nil {
		MyLog.Error("读取配置文件错误 " + err.Error())
	}

	// 赋初始值
	conf := configuration{defaultConsole}

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		MyLog.Error("解析配置文件错误", err.Error())
	}
	return conf
}

/**
	设置基础目录
 */
func SetBaseDir(dir string) {
	baseDir = dir
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

/**
	默认配置文件
 */
func defaultConfPath() string {
	return baseDir + "/config.yml"
}

/**
	默认目录
 */
func defaultConsolePath() []string {
	var path []string
	path = append(path, baseDir + "/console/")	// 项目
	return path
}

/**
	set
 */
func SetConsolePath(path []string) {
	for _, p := range path {
		consolePath = append(consolePath, p)
	}
}