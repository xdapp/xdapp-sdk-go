package register

import (
	"fmt"
	"path"
	"io/ioutil"
	"strings"
	"github.com/ddliu/go-httpclient"
	"log"
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

var (
	updateChan = make(chan RespFields, 5)
	removeChan = make(chan RespFields, 5)
)

/**
	console页面文件列表
 */
var vueFileList []string

/**
	同步Console页面文件
 */
func (reg *SRegister) ConsolePageSync() {

	MyLog.Debug("前端文件目录：" + JsonEncode(consolePath))

	defer func(){
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()

	list 	:= reg.checkConsolePageDiff()
	change 	:= list[0]
	remove 	:= list[1]

	MyLog.Debug("待修改文件列表:" + JsonEncode(change))
	MyLog.Debug("待删除文件列表:" + JsonEncode(remove))

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
			go reg.updateConsolePage(f, fullPath)
		}
	}

	// 删除文件
	for _, rFile := range remove {
		go reg.deleteConsolePage(rFile)
	}

	for {
		select {
		case update := <-updateChan:
			if update.Status != "ok" {
				panic("RPC服务更新文件, 接口报错:" + update.Msg)
			}
		case remove := <-removeChan:
			if remove.Status != "ok" {
				panic("RPC服务删除文件, 接口报错:" + remove.Msg)
			}
		}
	}
}

/**
	检查页面差异
 */
func (reg *SRegister) checkConsolePageDiff() [][]string  {

	if reg.ServiceData["pageServer"] == nil {
		panic("缺少请求同步地址！")
	}

	host := reg.ServiceData["pageServer"]["host"]
	key := reg.getKey()
	list := JsonEncode(getAllConsolePages())
	sign := Md5(fmt.Sprintf("%s.%s.%s", IntToStr(Time()), list, key))

	url := fmt.Sprintf("%scheck/%s/%s?time=%s&sign=%s", host, reg.App, reg.Name, IntToStr(Time()), sign)

	MyLog.Info("获取console前端文件: " + list)
	MyLog.Debug("获取ServiceData: " + JsonEncode(reg.ServiceData))

	response, err := httpclient.WithOption(httpclient.OPT_TIMEOUT, timeOut).Post(url, map[string]string{
		"list": list,
	})
	if err != nil {
		panic("检查页面请求执行url错误" + err.Error())
	}

	body, _ := ioutil.ReadAll(response.Body)
	MyLog.Debug("检查页面curl返回：" +  string(body))

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
func (reg *SRegister) updateConsolePage (file string, fullPath string) {

	host := reg.getHost()
	key := reg.getKey()
	app := reg.GetApp()
	name := reg.GetName()
	hash := Md5File(fullPath)
	timeStr := IntToStr(Time())
	sign := Md5(fmt.Sprintf("%s.%s.%s.%s", timeStr, file, hash, key))

	url := fmt.Sprintf("%sup/%s/%s/%s?time=%s&hash=%s&sign=%s", host, app, name, file, timeStr, hash, sign)
	MyLog.Debug("更新请求地址：" +  url)

	b, err := ioutil.ReadFile(fullPath)
	if err != nil {
		fmt.Print(err)
	}

	response, err := httpclient.WithOption(httpclient.OPT_TIMEOUT, timeOut).Put(url, strings.NewReader(string(b)))
	if err != nil {
		panic("更新文件执行curl错误" + err.Error())
	}

	body, _ := ioutil.ReadAll(response.Body)
	MyLog.Debug("更新文件 - " + file +  " ，curl返回：" +  string(body))

	resp := RespFields{}
	JsonDecode(string(body), &resp)
	updateChan<-resp
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
	MyLog.Debug("删除请求地址：" +  url)

	response, err := httpclient.WithOption(httpclient.OPT_TIMEOUT, timeOut).Get(url, map[string]string{})
	if err != nil {
		panic("删除文件执行curl错误" + err.Error())
	}

	body, _ := ioutil.ReadAll(response.Body)
	MyLog.Debug("删除文件 - " + file + "，curl返回：" +  string(body))

	resp := RespFields{}
	JsonDecode(string(body), &resp)
	removeChan<-resp
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