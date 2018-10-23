package service

import "reflect"

type IRegister interface {
	GetApp() string
	GetKey() string
	GetName() string
	GetVersion() string
	GetFunctions() []string
	SetRegSuccess(status bool)
	SetServiceData(data map[string]map[string]string)
	CloseClient()
	ConsolePageSync()
	RpcCall(name string, args []reflect.Value, namespace string, cfg map[string]uint32) interface{}
	ILogger
}

/**
logger
*/
type ILogger interface {
	Info(arg0 interface{}, args ...interface{})
	Debug(arg0 interface{}, args ...interface{})
	Warn(arg0 interface{}, args ...interface{})
	Error(arg0 interface{}, args ...interface{})
}
