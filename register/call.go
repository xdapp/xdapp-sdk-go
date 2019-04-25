package register

import (
	"time"
	"reflect"
	"strings"
	"encoding/binary"
	"github.com/leesper/tao"
)

var (
	receiveBuffer  map[string][]byte
	receiveChanMap = make (map[string]chan interface{})
)

const (
	RPC_CALL_WORKID    = 0	 // rpc workId (PHP版本对应进程id)
	RPC_CALL_TimeOut   = 10	 // rpc 请求超时时间
	RPC_CLEAR_BUF_TIME = 30	 // rpc 清理数据时间
)

type RpcClient struct {
	ServiceId uint32
	AdminId   uint32
	TimeOut   uint32
	NameSpace string
}

func NewRpcClient(c RpcClient) *RpcClient {
	if c.TimeOut == 0 {
		c.TimeOut = RPC_CALL_TimeOut
	}
	return &c
}

func (c *RpcClient) SetAdminId(AdminId uint32) {
	c.AdminId = AdminId
}

func (c *RpcClient) SetTimeOut(TimeOut uint32) {
	c.TimeOut = TimeOut
}

func (c *RpcClient) SetNameSpace(NameSpace string) {
	c.NameSpace = NameSpace
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
func (c *RpcClient) Call(name string, args []reflect.Value) interface{} {
	if c.NameSpace != "" {
		c.NameSpace = strings.TrimSuffix(c.NameSpace, "_") + "_"
		name = c.NameSpace + name
	}

	body := rpcEncode(name, args)
	requestId := uint32 (requestId.GetAndIncrement())

	var rpcContext = make([]byte, 2)
	binary.BigEndian.PutUint16(rpcContext, uint16(RPC_CALL_WORKID))
	prefix := Prefix{0,1,uint32(HEADER_LENGTH + len(rpcContext) + len(body))}
	header := Header{0,c.ServiceId,requestId,c.AdminId,uint8(len(rpcContext))}
	sendRequest(Request{prefix, header, rpcContext, body})

	time.Sleep(10 * time.Millisecond)
	timeId := Conn.RunAfter(time.Duration(c.TimeOut) * time.Second, func(i time.Time, closer tao.WriteCloser) {
		Logger.Info("Cancel the context")
	})
	defer Conn.CancelTimer(timeId)

	select {
	case result := <-receiveChanMap[IntToStr(requestId)]:
		delete(receiveChanMap, IntToStr(requestId))
		return result
	}
}

// rpc 请求返回
func sendRpcReceive(flag byte, header Header, body[]byte) {

	reqId := IntToStr(header.RequestId)
	finish := (flag & FLAG_FINISH) == FLAG_FINISH

	if finish == false {
		receiveBuffer[reqId] = body
		// 30秒后清理数据
		Conn.RunAt(time.Now().Add(RPC_CLEAR_BUF_TIME * time.Second), func(i time.Time, closer tao.WriteCloser) {
			delete(receiveBuffer, reqId)
		}); return
	} else if receiveBuffer[reqId] != nil {
		body = BytesCombine(receiveBuffer[reqId], body)
		delete(receiveBuffer, reqId)
	}

	if resp, error := rpcDecode(body); error != "" {
		Logger.Warn(error)
	} else {
		receiveChanMap[reqId] = make (chan interface{})
		receiveChanMap[reqId]<-resp
	}
}