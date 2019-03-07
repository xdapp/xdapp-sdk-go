package register

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"github.com/leesper/tao"
	"encoding/binary"
)

var (
	receiveBuffer  map[string][]byte
	callReceiveChanMap = make (map[string]chan interface{})
)

const (
	RPC_CALL_WORKID    = 0	 // rpc workId (PHP版本对应进程id)
	RPC_CALL_TIMEOUT   = 10	 // rpc 请求超时时间
	RPC_CLEAR_BUF_TIME = 30	 // rpc 清理数据时间
)

type RpcCall struct {
	serviceId uint32
	adminId uint32
	nameSpace string
	timeOut uint32
}

func NewRpcCall(c RpcCall) *RpcCall {
	if c.timeOut == 0 {
		c.timeOut = RPC_CALL_TIMEOUT
	}
	return &c
}

func (c *RpcCall) SetAdminId(adminId uint32) {
	c.adminId = adminId
}

func (c *RpcCall) SetTimeOut(timeOut uint32) {
	c.timeOut = timeOut
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
# 其中 Header 部分包括RPC服务连接关闭，等待重新连接
#
# AppId     | 服务ID      | rpc请求序号  | 管理员ID      | 自定义信息长度
# ----------|------------|------------|-------------|-----------------
# AppId     | ServiceId  | RequestId  | AdminId     | ContextLength
# 4         | 4          | 4          | 4           | 1
# N         | N          | N          | N           | C
*/
func (c *RpcCall) Call(name string, args []reflect.Value) interface{} {
	if c.nameSpace != "" {
		c.nameSpace = strings.TrimSuffix(c.nameSpace, "_") + "_"
		name = c.nameSpace + name
	}
	body := rpcEncode(name, args)

	// 唯一id
	reqId    := uint32(getGID())
	fmt.Println("gid is", reqId)

	// PHP版本对应进程id
	var rpcContext []byte = make([]byte, 2)
	binary.BigEndian.PutUint16(rpcContext, uint16(RPC_CALL_WORKID))

	prefix := Prefix{
		0,
		1,
		getRequestLength(rpcContext, body),
	}
	header := Header{
		0,
		c.serviceId,
		reqId,
		c.adminId,
		uint8(len(rpcContext)),
	}
	sendRequest(Request{prefix, header, rpcContext, body})

	time.Sleep(10 * time.Millisecond)

	timeId := Conn.RunAfter(time.Duration(c.timeOut) * time.Second, func(i time.Time, closer tao.WriteCloser) {
		Logger.Info("Cancel the context")
	})
	defer Conn.CancelTimer(timeId)

	reqIdStr := IntToStr(reqId)
	select {
	case result := <-callReceiveChanMap[reqIdStr]:
		delete(callReceiveChanMap, reqIdStr)
		Logger.Info(fmt.Sprintf("数量 %d", len(callReceiveChanMap)))
		return result
	}
}

// rpc 请求返回
func sendRpcReceive(flag byte, header Header, body[]byte) {

	id := IntToStr(header.RequestId)
	finish := (flag & FLAG_FINISH) == FLAG_FINISH

	if finish == false {
		receiveBuffer[id] = body
		// 30秒后清理数据
		Conn.RunAt(time.Now().Add(RPC_CLEAR_BUF_TIME * time.Second), func(i time.Time, closer tao.WriteCloser) {
			delete(receiveBuffer, id)
		}); return
	} else if receiveBuffer[id] != nil {
		body = BytesCombine(receiveBuffer[id], body)
		delete(receiveBuffer, id)
	}

	if resp, error := rpcDecode(body); error != "" {
		Logger.Warn(error)
	} else {
		callReceiveChanMap[id] = make (chan interface{})
		callReceiveChanMap[id]<-resp
	}
}

// 测试rpc
func TestRpcPing() {
	now := time.Now().Unix()
	args :=[]reflect.Value {reflect.ValueOf(now)}

	rpc := NewRpcCall(RpcCall{
		nameSpace: "test",
	})
	result := rpc.Call("ping", args)
	fmt.Println("rpc返回结果", result, now)
}