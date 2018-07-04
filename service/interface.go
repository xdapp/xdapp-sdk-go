package service

/**
	调用register接口
	https://wangzhezhe.github.io/2017/02/06/import-cycle/ (互相调用)
 */
type IRegister interface {
	GetApp() string
	GetName() string
	GetKey() string
	SetRegSuccess(status bool)
	SetServiceData(data map[string]map[string]string)
	CloseClient()
	ConsolePageSync()
	Logger
}

/**
	logger
 */
type Logger interface {
	Info(arg0 interface{}, args ...interface{})
	Debug(arg0 interface{}, args ...interface{})
	Warn(arg0 interface{}, args ...interface{})
	Error(arg0 interface{}, args ...interface{})
}


type NormalService struct {
	register IRegister
}

/**
	rpc 工厂建立
 */
func NewService(register IRegister) *NormalService {
	return &NormalService{register: register }
}

func (service *NormalService) getApp() string {
	return service.register.GetApp()
}

func (service *NormalService) getName() string {
	return service.register.GetName()
}

func (service *NormalService) getKey() string {
	return service.register.GetKey()
}