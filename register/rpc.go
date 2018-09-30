package register

import (
	"reflect"
	"github.com/hprose/hprose-golang/rpc"
	"github.com/hprose/hprose-golang/io"
	"runtime"
)

var (
	RpcService *rpc.TCPService		// rpc 服务
	RpcContext *rpc.SocketContext	// 上下文
)

func init() {
	RpcService = rpc.NewTCPService()
	RpcContext = new(rpc.SocketContext)
	RpcContext.InitServiceContext(RpcService)
}

/**
屏蔽列表输出
*/
func DoFunctionList() string {
	return "Fa{}z"
}

/**
执行结果
*/
func RpcHandle(data []byte) []byte {
	return RpcService.Handle(data, RpcContext)
}

func AddFunction(name string, function interface{}, option ...rpc.Options) {
	RpcService.AddFunction(name, function, option...)
}

func AddInstanceMethods(obj interface{}, namespace string) {
	RpcService.AddInstanceMethods(obj, rpc.Options{NameSpace: namespace})
}

func rpcEncode(name string, args []reflect.Value) []byte {
	writer := io.NewWriter(false)
	writer.WriteByte(io.TagCall)
	writer.WriteString(name)
	writer.Reset()
	writer.WriteSlice(args)
	writer.WriteByte(io.TagEnd)
	return writer.Bytes()
}

func (reg *SRegister) RpcCall(name string, args []reflect.Value) interface{} {

	go func(name string, args []reflect.Value) interface{} {

		id := uint32(1)
		workerId := uint16(1)
		custom := Pack(workerId)
		body   := rpcEncode(name, args)
		length := uint32(HEADER_LENGTH + len(custom) + len(body))

		request := &SRequest{0, 1, length,SHeader{0, 1, id, 1,uint8(len(custom))}}

		call := BytesCombine(Pack(request), custom, body)
		reg.Client.Send(call)

		runtime.Gosched()
		return <-rpcCallChan
	}(name, args)

	return <-rpcCallChan
}

func rpcCall1() {
	//ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
	//defer cancel()
}

//// NewContext returns a new Context carrying userIP.
//func NewContext(ctx context.Context, userIP net.IP) context.Context {
//	return context.WithValue(ctx, userIPKey, userIP)
//}
//
//// FromContext extracts the user IP address from ctx, if present.
//func FromContext(ctx context.Context) (net.IP, bool) {
//	// ctx.Value returns nil if ctx has no value for the key;
//	// the net.IP type assertion returns ok=false for nil.
//	userIP, ok := ctx.Value(userIPKey).(net.IP)
//	return userIP, ok
//}