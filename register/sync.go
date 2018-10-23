package register

import (
	"fmt"
	"log"
	"sync"
	"path"
	"strings"
	"io/ioutil"
	"github.com/ddliu/go-httpclient"
)

/**
返回结构体
*/
type RespFields struct {
	Status string     // 状态
	Msg    string     // 消息
	List   [][]string // 列表
}

/**
超时
*/
const timeOut = 5

var wg sync.WaitGroup

/**
console页面文件列表
*/
var vueFileList []string

/**
前端页面目录
  */
var consolePath []string

func setConsolePath(path []string) {
	var descPath []string
	// 校验前端目录
	for _, p := range path {
		if !IsExist(p) {
			continue
		}
		descPath = append(descPath, p)
	}
	consolePath = descPath
}

/**
同步Console页面文件
*/
func (reg *SRegister) ConsolePageSync() {

	Logger.Debug("前端文件目录：" + JsonEncode(consolePath))

	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()

	list   := reg.checkConsolePageDiff()
	change := list[0]
	remove := list[1]

	Logger.Debug("待修改文件列表:" + JsonEncode(change))
	Logger.Debug("待删除文件列表:" + JsonEncode(remove))

	// 更新的文件
	for _, f := range change {
		fullPath := ""
		for _, conDir := range consolePath {
			if PathExist(conDir + f) {
				fullPath = conDir + f
				break
			}
		}
		if fullPath != "" {
			wg.Add(1)
			go func(file string, fullPath string) {
				defer wg.Done()
				reg.updateConsolePage(file, fullPath)
			}(f, fullPath)
		}
	}

	// 删除文件
	for _, rFile := range remove {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			reg.deleteConsolePage(file)
		}(rFile)
	}
	wg.Wait()
}

/**
检查页面差异
*/
func (reg *SRegister) checkConsolePageDiff() [][]string {

	if reg.ServiceData["pageServer"] == nil {
		panic("缺少请求同步地址！")
	}

	time := IntToStr(Time())
	list := JsonEncode(getAllConsolePages())
	sign := Md5(fmt.Sprintf("%s.%s.%s", time, list, reg.getKey()))
	url  := reg.getUrl("check", sign, time, map[string]string{})

	Logger.Info("获取console url: " + url)
	Logger.Info("获取console前端文件: " + list)
	Logger.Debug("获取ServiceData: " + JsonEncode(reg.ServiceData))

	response, err := httpclient.WithOption(httpclient.OPT_TIMEOUT, timeOut).Post(url, map[string]string{
		"list": list,
	})
	if err != nil {
		panic("检查页面请求执行url错误" + err.Error())
	}

	body, _ := ioutil.ReadAll(response.Body)
	Logger.Debug("检查页面curl返回：" + string(body))

	resp := RespFields{}
	JsonDecode(string(body), &resp)
	if resp.Status != "ok" {
		panic("RPC服务同步page文件获取列表返回: " + string(body))
	}

	return resp.List
}

/**
执行更新文件
*/
func (reg *SRegister) updateConsolePage(file string, fullPath string) {

	hash := Md5File(fullPath)
	time := IntToStr(Time())
	sign := Md5(fmt.Sprintf("%s.%s.%s.%s", time, file, hash, reg.getKey()))
	url  := reg.getUrl("upload", sign, time, map[string]string{"file":file, "hash": hash})

	b, err := ioutil.ReadFile(fullPath)
	if err != nil {
		fmt.Print(err)
	}

	response, err := httpclient.WithOption(httpclient.OPT_TIMEOUT, timeOut).Put(url, strings.NewReader(string(b)))
	if err != nil {
		panic("更新文件执行curl错误" + err.Error())
	}

	body, _ := ioutil.ReadAll(response.Body)
	Logger.Debug("更新文件 - " + file + " ，curl返回：" + string(body))

	resp := RespFields{}
	JsonDecode(string(body), &resp)
	if resp.Status != "ok" {
		panic("RPC服务更新文件, 接口报错:" + resp.Msg)
	}
}

/**
执行删除文件
*/
func (reg *SRegister) deleteConsolePage(file string) {

	time := IntToStr(Time())
	sign := Md5(fmt.Sprintf("%s.%s.%s", time, file, reg.getKey()))

	url  := reg.getUrl("remove", sign, time, map[string]string{"file":file})

	Logger.Debug("删除请求地址：" + url)

	response, err := httpclient.WithOption(httpclient.OPT_TIMEOUT, timeOut).Get(url, map[string]string{})
	if err != nil {
		panic("删除文件执行curl错误" + err.Error())
	}

	body, _ := ioutil.ReadAll(response.Body)
	Logger.Debug("删除文件 - " + file + "，curl返回：" + string(body))

	resp := RespFields{}
	JsonDecode(string(body), &resp)
	if resp.Status != "ok" {
		panic("RPC服务删除文件, 接口报错:" + resp.Msg)
	}
}

/**
获取所有Console用到的页面列表
*/
func getAllConsolePages() map[string]string {

	list := make(map[string]string)
	for _, dir := range consolePath {

		dirLen := len(dir)
		loopFindAllFile(dir)

		for _, f := range vueFileList {
			if ext := path.Ext(f); ext == ".tpl" || ext == ".vue" {
				fileName := Substr(f, dirLen, len(f)-dirLen)
				list[fileName] = Md5File(f)
			}
		}
	}
	return list
}

func (reg *SRegister) getUrl(action string, sign string, time string, ext map[string]string) string {
	var url string
	switch action {
		case "check":
			url = fmt.Sprintf("%scheck/%s/%s?time=%s&sign=%s",
				reg.getHost(),
				reg.GetApp(),
				reg.GetName(), time, sign)
		case "upload":
			url = fmt.Sprintf("%sup/%s/%s/%s?time=%s&hash=%s&sign=%s",
				reg.getHost(),
				reg.GetApp(),
				reg.GetName(), ext["file"], time, ext["hash"], sign)
		case "remove":
			url = fmt.Sprintf("%srm/%s/%s/%s?time=%s&sign=%s",
				reg.getHost(),
				reg.GetApp(),
				reg.GetName(),
				ext["file"], time, sign)
		default:
	}

	Logger.Debug(action + "url: " + url)
	return url
}

/**
获取当前文件夹下全部文件
*/
func loopFindAllFile(folder string) {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {

		folder = strings.TrimRight(folder, "/")
		if file.IsDir() {
			loopFindAllFile(folder + "/" + file.Name())
		} else {
			vueFileList = append(vueFileList, folder+"/"+file.Name())
		}
	}
}
