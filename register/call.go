package register

import (
	"encoding/binary"
	"github.com/leesper/tao"
	"reflect"
	"strings"
	"time"
)

const (
	RpcCallWorkId    = 0	 // rpc workId (PHP版本对应进程id)
	RpcCallTimeout   = 10	 // rpc 请求超时时间
	RpcClearBufTime = 30	 // rpc 清理数据时间
)

type rpcClient struct {
	Conn     *tao.ClientConn
	ServiceId uint32
	AdminId   uint32
	TimeOut   uint32
	NameSpace string
}

func NewRpcClient(conn *tao.ClientConn, serviceId uint32, adminId uint32, timeOut uint32,  nameSpace string) *rpcClient {
	if timeOut == 0 {
		timeOut = RpcCallTimeout
	}
	return &rpcClient{
		Conn:      conn,
		ServiceId: serviceId,
		AdminId:   adminId,
		TimeOut:   timeOut,
		NameSpace: nameSpace,
	}
}

func (c *rpcClient) SetAdminId(AdminId uint32) {
	c.AdminId = AdminId
}

func (c *rpcClient) SetTimeOut(TimeOut uint32) {
	c.TimeOut = TimeOut
}

func (c *rpcClient) SetNameSpace(NameSpace string) {
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
func (c *rpcClient) Call(name string, args []reflect.Value) interface{} {
	if c.NameSpace != "" {
		c.NameSpace = strings.TrimSuffix(c.NameSpace, "_") + "_"
		name = c.NameSpace + name
	}

	body := rpcEncode(name, args)
	requestId := uint32 (requestId.GetAndIncrement())

	var rpcContext = make([]byte, 2)
	binary.BigEndian.PutUint16(rpcContext, uint16(RpcCallWorkId))
	prefix := Prefix{0,1,uint32(HeaderLength + len(rpcContext) + len(body))}
	header := Header{0,c.ServiceId,requestId,c.AdminId,uint8(len(rpcContext))}
	sendRequest(c.Conn, Request{prefix, header, rpcContext, body})

	time.Sleep(10 * time.Millisecond)
	timeId := c.Conn.RunAfter(time.Duration(c.TimeOut) * time.Second, func(i time.Time, closer tao.WriteCloser) {
		Logger.Info("Cancel the context")
	})
	defer c.Conn.CancelTimer(timeId)

	reqId := IntToStr(requestId)
	select {
	case result := <-receiveChanMap[reqId]:
		delete(receiveChanMap, reqId)
		return result
	}
}