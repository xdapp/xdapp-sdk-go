package register

import (
	"fmt"
	"bytes"
	"time"
	"reflect"
	"strings"
	"github.com/leesper/tao"
	"github.com/hprose/hprose-golang/rpc"
	"github.com/hprose/hprose-golang/io"
)

var (
	RpcService *rpc.TCPService		 // rpc 服务
	RpcContext *rpc.SocketContext	 // 上下文
	receiveBuffer  map[string][]byte // 接收的rpc请求数据
)

const (
	callTimeout = 5
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

/*
# 将请求发送给RPC连接进程

# 标识   | 版本    | 长度    | 头信息       | 自定义内容    |  正文
# ------|--------|---------|------------|-------------|-------------
# Flag  | Ver    | Length  | Header     | Context      | Body
# 1     | 1      | 4       | 17         | 默认0，不定   | 不定
# C     | C      | N       |            |             |
#
#
# 其中 Header 部分包括
#
# AppId     | 服务ID      | rpc请求序号  | 管理员ID      | 自定义信息长度
# ----------|------------|------------|-------------|-----------------
# AppId     | ServiceId  | RequestId  | AdminId     | ContextLength
# 4         | 4          | 4          | 4           | 1
# N         | N          | N          | N           | C
*/
func RpcCall(name string, args []reflect.Value, namespace string, cfg map[string]uint32) interface{} {

	if _, ok := cfg["serviceId"]; !ok {
		cfg["serviceId"] = 0
	}
	if _, ok := cfg["adminId"]; !ok {
		cfg["adminId"] = 0
	}

	// 唯一id
	id    := uint32(getGID())
	idStr := IntToStr(id)
	fmt.Println("gid is  ", id)

	// PHP版本对应进程id
	var writer = new(bytes.Buffer)
	pack(writer, uint16(0))
	rpcContext := writer.Bytes()

	namespace = strings.TrimSuffix(namespace, "_") + "_"
	name = namespace + name

	socketSendChan<-setRequest(0, 1, Header{
		0,
		cfg["serviceId"],
		id,
		cfg["adminId"],
		uint8(len(rpcContext)),
	}, rpcContext, rpcEncode(name, args))

	time.Sleep(10*time.Millisecond)
	timeId := Conn.RunAfter(callTimeout*time.Second, func(i time.Time, closer tao.WriteCloser) {
		fmt.Println("Cancel the context")
	})
	defer Conn.CancelTimer(timeId)

	select {
	case callReturn := <-rpcCallRespMap[idStr]:
		delete(rpcCallRespMap, idStr)
		fmt.Println("数量", len(rpcCallRespMap))
		return callReturn
	}
}

/**
rpc 请求返回
 */
func rpcReceive(flag byte, header Header, body[]byte) {

	id := IntToStr(header.RequestId)
	finish := (flag & FLAG_FINISH) == FLAG_FINISH

	if finish == false {
		receiveBuffer[id] = body
		// 30秒后清理数据
		Conn.RunAt(time.Now().Add(30 * time.Second), func(i time.Time, closer tao.WriteCloser) {
			delete(receiveBuffer, id)
		}); return
	} else if receiveBuffer[id] != nil {
		body = BytesCombine(receiveBuffer[id], body)
		delete(receiveBuffer, id)
	}

	if result, error := rpcDecode(body); error != "" {
		Logger.Warn(error)
	} else {
		fmt.Println("解析rpc 请求", id, result)
		rpcCallRespMap[id] = make (chan interface{})
		rpcCallRespMap[id]<-result
	}
}

/**
rpc序列化
 */
func rpcEncode(name string, args []reflect.Value) []byte {
	writer := io.NewWriter(false)
	writer.WriteByte(io.TagCall)
	writer.WriteString(name)
	writer.Reset()
	writer.WriteSlice(args)
	writer.WriteByte(io.TagEnd)
	return writer.Bytes()
}

/**
rpc反序列化
 */
func rpcDecode(data []byte) (interface{}, string) {
	reader := io.AcquireReader(data, false)
	defer io.ReleaseReader(reader)
	tag, _ := reader.ReadByte()
	switch tag {
	case io.TagResult:
		var e interface{}
		reader.Unserialize(&e)
		return e, ""
	case io.TagError:
		return nil, "RPC 系统调用 Agent 返回错误信息: " + reader.ReadString()
	default:
		return nil, "RPC 系统调用收到一个未定义的方法返回: " + string(tag) + reader.ReadString()
	}
}