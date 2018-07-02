package service

/**
	调用register接口
	https://wangzhezhe.github.io/2017/02/06/import-cycle/ (互相调用)
 */
type RegisterInterFace interface {
	GetApp() string
	GetName() string
	GetKey() string
	SetRegSuccess(status bool)
	SetServiceData(data map[string]map[string]string)
	CloseClient()
	ConsolePageSync()
}

type Service struct {
	RegisterFace RegisterInterFace
}

/**
	rpc 工厂建立
 */
func NewService(RegisterFace RegisterInterFace) *Service {
	return &Service{RegisterFace: RegisterFace }
}

func (service *Service) getApp() string {
	return service.RegisterFace.GetApp()
}

func (service *Service) getName() string {
	return service.RegisterFace.GetName()
}

func (service *Service) getKey() string {
	return service.RegisterFace.GetKey()
}