package register

import "reflect"

func (reg *SRegister) GetApp() string {
	return reg.Console.App
}

func (reg *SRegister) GetName() string {
	return reg.Console.Name
}
func (reg *SRegister) GetVersion() string {
	return reg.Version
}
func (reg *SRegister) GetKey() string {
	return reg.Console.Key
}

func (reg *SRegister) SetRegSuccess(status bool) {
	reg.RegSuccess = status
}

func (reg *SRegister) SetServiceData(data map[string]map[string]string) {
	reg.ServiceData = data
}

func (reg *SRegister) GetFunctions() []string {
	return RpcService.MethodNames
}

func (reg *SRegister) CloseClient() {
	reg.Conn.Close()
}

func (reg *SRegister) Info(arg0 interface{}, args ...interface{}) {
	reg.Logger.Info(arg0, args...)
}

func (reg *SRegister) Debug(arg0 interface{}, args ...interface{}) {
	reg.Logger.Debug(arg0, args...)
}

func (reg *SRegister) Warn(arg0 interface{}, args ...interface{}) {
	reg.Logger.Warn(arg0, args...)
}

func (reg *SRegister) Error(arg0 interface{}, args ...interface{}) {
	reg.Logger.Error(arg0, args...)
}

func (reg *SRegister) RpcCall(name string, args []reflect.Value, namespace string, cfg map[string]uint32) interface{} {
	return RpcCall(name, args, namespace, cfg)
}
