package register

import (
	"fmt"
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

var (
	conf       configuration // 配置
	baseDir    = defaultBaseDir()
	configPath = defaultConfPath()
)

/**
	设置配置
 */
func LoadConfig() {

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}
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