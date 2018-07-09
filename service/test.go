package service

/**
	测试增加扩展类
 */
type TestService struct {
	name string
}
func NewTestService(name string) *TestService{
	return &TestService{name}
}
func (service *TestService) Say() string {
	return service.name
}
