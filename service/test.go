package service

// 测试增加扩展类
type TestService struct {
	Name string
}

func (service *TestService) Say() string {
	return service.Name
}
