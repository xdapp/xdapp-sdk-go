package register

import (
	"strings"
	"reflect"
)

/**
	成功服务列表
 */
var sucService = make(map[string]string)

/**
	获取可执行的function
  */
func LoadService(name string, service interface{}) {

	t := reflect.TypeOf(service)
	v := reflect.ValueOf(service)

	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		mv := v.MethodByName(m.Name) 	//获取对应的方法
		if !mv.IsValid() {            	//判断方法是否存在
			MyLog.Error("Func Hello not exist")
			continue
		}

		// 注册rpc方法
		rpcName := name + "_" + strings.ToLower(m.Name)
		if sucService[rpcName] != ""  {
			MyLog.Error("RPC服务已经存在 " + rpcName + ", 已忽略 ")
		}

		MyLog.Debug("增加RPC方法： " + rpcName)

		MyRpc.AddFunction(rpcName, mv)

		sucService[rpcName] = "{" + name + "}" +  rpcName

		//args := []reflect.Value{reflect.ValueOf(m)} //初始化传入等参数，传入等类型只能是[]reflect.Value类型
		//res := mv.Call(args)
	}
}