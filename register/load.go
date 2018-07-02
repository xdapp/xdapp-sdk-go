package register

import (
	"../service"
	"../util"
	"strings"
	"reflect"
)

/**
	加载rpc服务
 */
func (reg *RegisterData) LoadService() {

	dirArr := GetPath("service")

	list := make(map[string]string)

	for _, dir := range dirArr {
		loadServiceByPath(&list, dir, "")
	}

	Debug(list)
	Debug("获取到RPC服务", list)

	// add rpc func
	for name, _ := range list {
		isSysCall := strings.ToLower(name) == "sys" || strings.ToLower(name) == "sys_"

		var newService interface{}
		newService = reg.NewService()
		if isSysCall {
			newService = reg.NewSysService()
		}
		AddRpcFunction(name, newService)
	}
}

/**
	根据路径加载服务代码
 */
func loadServiceByPath(list *map[string]string, dir string, prefix string)  {

	for _, f := range util.FindAllFiles(dir) {

		base, ext := util.GetFileInfo(f)
		name := strings.ToLower(base)

		if ext != ".go" {
			continue
		}

		name = string([]rune(name)[:len(name)-3])

		(*list)[prefix+name] = f
	}
}

/**
	获取可执行的function
  */
func AddRpcFunction(name string, avail interface{}) {

	t := reflect.TypeOf(avail)
	v := reflect.ValueOf(avail)

	success := make(map[string]string)

	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		mv := v.MethodByName(m.Name) 	//获取对应的方法
		if !mv.IsValid() {            	//判断方法是否存在
			Error("Func Hello not exist")
			continue
		}

		// 注册rpc方法
		rpcName := name + "_" + strings.ToLower(m.Name)
		if success[rpcName] != ""  {
			Error("RPC服务已经存在 " + rpcName + ", 已忽略, File: ")
		}

		Debug("增加RPC方法： " + rpcName)

		myRpc.service.AddFunction(rpcName, mv)

		success[rpcName] = "{" + name + "}" +  rpcName

		//args := []reflect.Value{reflect.ValueOf(m)} //初始化传入等参数，传入等类型只能是[]reflect.Value类型
		//res := mv.Call(args)
	}
}

/**
	工厂注册
 */
func (reg *RegisterData) NewService() *service.Service {
	return service.NewService(service.RegisterInterFace(reg))
}

func (reg *RegisterData) NewSysService() *service.SysService {
	return service.NewSysService(service.RegisterInterFace(reg))
}