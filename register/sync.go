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
	同步Console页面文件
 */
func (reg *RegisterData) ConsolePageSync() {

	Debug("doing ConsolePageSync")

	dirArr 	:= GetPath("console")
	list 	:= reg.checkConsolePageDiff()
	change 	:= list[0]
	remove 	:= list[1]

	Debug("待修改文件列表", change)
	Debug("待删除文件列表", remove)

	reg.getUpdateFile(change, dirArr)

	// 删除文件
	for _, rFile := range remove {
		reg.deleteConsolePage(rFile)
	}
}

/**
	检查页面差异
 */
func (reg *RegisterData) checkConsolePageDiff() [][]string  {
	// pageServer

	if reg.ServiceData["pageServer"] == nil {
		return nil
	}

	key     := reg.getKey()
	list 	:= getAllConsolePages()
	listStr := JsonEncode(list)
	timeStr := IntToStr(Time())

	fmt.Println("console view", listStr)

	sign 	:= Md5(fmt.Sprintf("%s.%s.%s", timeStr, listStr, key))

	fmt.Println("reg.ServiceData", reg.ServiceData)
	host 	:= reg.ServiceData["pageServer"]["host"]
	reqUrl 	:= fmt.Sprintf("%scheck/%s/%s?time=%s&sign=%s", host, reg.App, reg.Name, timeStr, sign)

	fmt.Println("请求的同步地址", reqUrl)

	rs := CurlRequest(reqUrl, "list=" + listStr)
	resp := RespFields{}
	JsonDecode(rs, &resp)

	if  resp.Status == "" {
		fmt.Println("RPC服务同步page文件获取列表信息解析json失败: " + rs)
	}

	if resp.Status != "ok" {
		fmt.Println("RPC服务同步page文件获取列表返回: " + rs)
	}
	fmt.Println("curl request", resp)

	return resp.List
}

/**
	获取待更新的文件
 */
func (reg *RegisterData) getUpdateFile(change []string, dirArr []string) {

	var fullPath string

	for _, f := range change {
		fullPath = ""
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
func (reg *RegisterData) updateConsolePage (file string, fullPath string) {
	Debug(fullPath)

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
		Debug("RPC服务成功上传console页面" + fullPath)
	} else {
		Debug("RPC服务上传page文件返回" + err.Error())
	}
}

/**
	执行删除文件
 */
func (reg *RegisterData) deleteConsolePage(file string) {
	key  := reg.getKey()
	host := reg.getHost()
	app  := reg.GetApp()
	name := reg.GetName()

	timeStr := IntToStr(Time())

	sign := Md5(fmt.Sprintf("%s.%s.%s", timeStr, file, key))
	url  := fmt.Sprintf("%srm/%s/%s/%s?time=%s&sign=%s", host, app, name, file, timeStr, sign)
	rs   := CurlRequest(url, "")

	resp := RespFields{}
	JsonDecode(rs, &resp)

	if  resp.Status == "" {
		Error("RPC服务删除文件" + file + "获取列表信息解析json失败: " + rs)
	}

	if resp.Status != "ok" {
		Error("RPC服务删除文件" + file  + "接口返回" + rs)
	}

	Debug("删除文件结果返回：", rs)
}


/**
	获取所有Console用到的页面列表
 */
func getAllConsolePages() map[string]string {

	list := make(map[string]string)

	dirArr := GetPath("console")
	for _, dir := range dirArr {

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