package service

type IRegister interface {
	GetApp() string
	GetFunctions() []string
	GetName() string
	GetVersion() string
	GetKey() string
	SetRegSuccess(status bool)
	SetServiceData(data map[string]map[string]string)
	CloseClient()
	ConsolePageSync()
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
