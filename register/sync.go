package register

import (
	"fmt"
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
	前端页面目录
 */
var consolePath = defaultConsolePath()

/**
	同步Console页面文件
 */
func (reg *SRegister) ConsolePageSync() {

	MyLog.Debug("同步Console页面文件 ")

	list 	:= reg.checkConsolePageDiff()
	change 	:= list[0]
	remove 	:= list[1]

	MyLog.Debug("待修改文件列表", change)
	MyLog.Debug("待删除文件列表", remove)
	MyLog.Debug("console 目录", consolePath)

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

	// pageServer
	if reg.ServiceData["pageServer"] == nil {
		return nil
	}

	key     := reg.getKey()
	list 	:= getAllConsolePages()
	listStr := JsonEncode(list)
	sign 	:= Md5(fmt.Sprintf("%s.%s.%s", IntToStr(Time()), listStr, key))

	MyLog.Info("console view", listStr)
	MyLog.Debug("reg.ServiceData", reg.ServiceData)

	host 	:= reg.ServiceData["pageServer"]["host"]
	reqUrl 	:= fmt.Sprintf("%scheck/%s/%s?time=%s&sign=%s", host, reg.App, reg.Name, IntToStr(Time()), sign)

	MyLog.Debug("请求的同步地址", reqUrl)

	rs := Request(reqUrl, "list=" + listStr)
	resp := RespFields{}
	JsonDecode(rs, &resp)

	if  resp.Status == "" {
		MyLog.Error("RPC服务同步page文件获取列表信息解析json失败: " + rs)
	}

	if resp.Status != "ok" {
		MyLog.Error("RPC服务同步page文件获取列表返回: " + rs)
	}
	MyLog.Debug("curl request", resp)

	return resp.List
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
	MyLog.Debug(fullPath)

	host 	:= reg.getHost()
	key     := reg.getKey()
	app 	:= reg.GetApp()
	name 	:= reg.GetName()

	hash 	:= Md5File(fullPath)
	timeStr := IntToStr(Time())
	sign    := Md5(fmt.Sprintf("%s.%s.%s.%s", timeStr, file, hash, key))

	reqUrl := fmt.Sprintf("%sup/%s/%s/%s?time=%s&hash=%s&sign=%s", host, app, name, file, timeStr, hash, sign)
	err := PostFile(fullPath, reqUrl)

	if err != nil {
		MyLog.Debug("RPC服务成功上传console页面" + fullPath)
	} else {
		MyLog.Debug("RPC服务上传page文件返回" + err.Error())
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
	rs   := Request(url, "")

	resp := RespFields{}
	JsonDecode(rs, &resp)

	if  resp.Status == "" {
		MyLog.Error("RPC服务删除文件" + file + "获取列表信息解析json失败: " + rs)
	}

	if resp.Status != "ok" {
		MyLog.Error("RPC服务删除文件" + file  + "接口返回" + rs)
	}

	MyLog.Debug("删除文件结果返回：", rs)
}


/**
	获取所有Console用到的页面列表
 */
func getAllConsolePages() map[string]string {

	list := make(map[string]string)

	for _, dir := range consolePath {

		files := FindAllFiles(dir)
		for _, f := range files {

			// 判断后缀名是tpl 、vue
			base, ext := GetFileInfo(f)
			if ext == ".tpl" || ext == ".vue" {
				list[base] = Md5File(f)
			}
		}
	}
	return list
}