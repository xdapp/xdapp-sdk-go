package register

import (
	"fmt"
	"path"
	"io/ioutil"
	"strings"
	"github.com/ddliu/go-httpclient"
)

/**
	返回结构体
 */
type RespFields struct {
	Status	string				// 状态
	Msg	    string				// 消息
	List    [][]string			// 列表
}

/**
	超时
 */
const timeOut = 5

/**
	console页面文件列表
 */
var vueFileList []string

/**
	同步Console页面文件
 */
func (reg *SRegister) ConsolePageSync() {
	MyLog.Debug("前端文件目录：" + JsonEncode(consolePath))

	list 	:= reg.checkConsolePageDiff()
	change 	:= list[0]
	remove 	:= list[1]

	MyLog.Debug("设置的console 目录: " + JsonEncode(consolePath))
	MyLog.Debug("待修改文件列表:" + JsonEncode(change))
	MyLog.Debug("待删除文件列表:" + JsonEncode(remove))

	reg.getUpdateFile(change, consolePath)
	// 删除文件
	for _, rFile := range remove {
		reg.deleteConsolePage(rFile)
	}
}

/**
	检查页面差异
 */
func (reg *SRegister) checkConsolePageDiff() [][]string  {

	if reg.ServiceData["pageServer"] == nil {
		MyLog.Error("缺少请求同步地址！")
	}

	key := reg.getKey()
	list := JsonEncode(getAllConsolePages())
	sign := Md5(fmt.Sprintf("%s.%s.%s", IntToStr(Time()), list, key))

	host := reg.ServiceData["pageServer"]["host"]
	url := fmt.Sprintf("%scheck/%s/%s?time=%s&sign=%s", host, reg.App, reg.Name, IntToStr(Time()), sign)

	MyLog.Info("获取console前端文件: " + list)
	MyLog.Debug("获取ServiceData: " + JsonEncode(reg.ServiceData))

	response, err := httpclient.WithOption(httpclient.OPT_TIMEOUT, timeOut).Post(url, map[string]string{
		"list": list,
	})
	if err != nil {
		MyLog.Error("curl执行错误" + err.Error())
	}

	result := resolveResponse(response)
	if result.Status != "ok" {
		MyLog.Error("RPC服务同步page文件获取列表返回: ", result)
	}

	return result.List
}

/**
	获取待更新的文件
 */
func (reg *SRegister) getUpdateFile(change []string, dirArr []string) {

	for _, f := range change {
		fullPath := ""
		for _, dir := range dirArr {
			if PathExist(dir + f) {
				fullPath = dir + f
				break
			}
		}
		if fullPath != "" {
			reg.updateConsolePage(f, fullPath)
		}
	}
}

/**
	执行更新文件
 */
func (reg *SRegister) updateConsolePage (file string, fullPath string) {
	host 	:= reg.getHost()
	key     := reg.getKey()
	app 	:= reg.GetApp()
	name 	:= reg.GetName()

	hash 	:= Md5File(fullPath)
	timeStr := IntToStr(Time())
	sign    := Md5(fmt.Sprintf("%s.%s.%s.%s", timeStr, file, hash, key))

	url := fmt.Sprintf("%sup/%s/%s/%s?time=%s&hash=%s&sign=%s", host, app, name, file, timeStr, hash, sign)

	MyLog.Debug("更新请求地址：" +  url)

	b, err := ioutil.ReadFile(fullPath)
	if err != nil {
		fmt.Print(err)
	}

	response, err := httpclient.WithOption(httpclient.OPT_TIMEOUT, timeOut).Put(url, strings.NewReader(string(b)))

	if err != nil {
		MyLog.Error("curl执行错误" + err.Error())
	}

	result := resolveResponse(response)
	if result.Status != "ok" {
		MyLog.Error("RPC服务更新文件：" + file  + " 接口报错：", result)
	}
}

/**
	执行删除文件
 */
func (reg *SRegister) deleteConsolePage(file string) {

	key  := reg.getKey()
	app  := reg.GetApp()
	name := reg.GetName()
	host := reg.getHost()
	time := IntToStr(Time())

	sign := Md5(fmt.Sprintf("%s.%s.%s", time, file, key))
	url  := fmt.Sprintf("%srm/%s/%s/%s?time=%s&sign=%s", host, app, name, file, time, sign)

	response, err := httpclient.WithOption(httpclient.OPT_TIMEOUT, timeOut).Get(url, map[string]string{})
	if err != nil {
		MyLog.Error("curl执行错误" + err.Error())
	}

	result := resolveResponse(response)
	if result.Status != "ok" {
		MyLog.Error("RPC服务删除文件" + file  + "接口返回" , result)
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
				fileName := Substr(f, dirLen, len(f) - dirLen)
				list[fileName] = Md5File(f)
			}
		}
	}
	return list
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
			vueFileList = append(vueFileList, folder + "/" + file.Name())
		}
	}
}

func resolveResponse(response *httpclient.Response) RespFields {

	body, _ := ioutil.ReadAll(response.Body)
	resp := RespFields{}
	JsonDecode(string(body), &resp)

	MyLog.Debug("curl结果返回：" +  string(body))
	return resp
}