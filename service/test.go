package service

/**
测试增加扩展类
*/
type Test struct {
	Name string
}

func (service *Test) Say() string {
	return service.Name
}
