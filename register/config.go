package register

import (
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"log"
	"strings"
	"os"
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
	设置配置
 */
func LoadConfig(filePath string) configuration {

	if !PathExist(filePath) {
		MyLog.Error("配置文件：" + filePath + "不存在！")
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		MyLog.Error("读取配置文件错误 " + err.Error())
	}

	// 赋初始值
	conf := configuration{
		console{defaultHost, defaultSSl, defaultApp, defaultName, defaultKey}}

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		MyLog.Error("解析配置文件错误", err.Error())
	}
	return conf
}

/**
	默认基础目录
 */
func DefaultBaseDir() string {
	dir, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

/**
	默认前端目录
 */
func defaultConsolePath() string {
	return DefaultBaseDir() + "/console/"
}

/**
	校验前端目录
 */
func checkConsolePath(path []string) []string {
	var descPath[]string
	for _, p := range path {
		if !IsExist(p) {
			continue
		}
		descPath = append(descPath, p)
	}
	return descPath
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}