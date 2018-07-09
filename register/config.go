package register

import (
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"log"
	"strings"
	"os"
	"fmt"
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
func LoadConfig(filePath string) (configuration, error) {

	if !PathExist(filePath) {
		return configuration{}, fmt.Errorf("配置文件:%s 不存在", filePath)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return configuration{}, fmt.Errorf("读取配置文件错误:%s", err.Error())
	}

	// 赋初始值
	conf := configuration{
		console{defaultHost, defaultSSl, defaultApp, defaultName, defaultKey}}

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return configuration{}, fmt.Errorf("解析配置文件错误:%s", err.Error())
	}
	return conf, nil
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