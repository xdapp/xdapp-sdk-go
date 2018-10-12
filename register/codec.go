package register

import (
	"bytes"
	"encoding/binary"
)

//  标识   | 版本    | 长度    | 头信息       | 自定义内容    |  正文
//  ------|--------|---------|------------|-------------|-------------
//  Flag  | Ver    | Length  | Header     | Context      | Body
//  1     | 1      | 4       | 17         | 默认0，不定   | 不定
//  C     | C      | N       |            |             |
//
//
//  其中 Header 部分包括
//
//  服务ID     | rpc请求序号  | 管理员ID      | 自定义信息长度
//  ----------|-------------|-------------|-----------------
//  ServiceId | RequestId   | AdminId     | ContextLength
//  4         | 4           | 4           | 1
//  N         | N           | N           | C

type SRequest struct {
	Flag    byte   // 标志位 成功 |= 4
	Ver     byte   // 版本
	Length  uint32 // 长度
	SHeader		   // 头信息
}

type SHeader struct {
	AppId         uint32
	ServiceId     uint32
	RequestId     uint32
	AdminId       uint32
	ContextLength byte
}

type SResponse struct {
	Flag    byte   // 标志位 成功 |= 4
	Ver     byte   // 版本
	Length  uint32 // 长度
}

const (
	PREFIX_LENGTH  = 6                             // Flag 1字节、 Ver 1字节、 Length 4字节、HeaderLength 1字节
	HEADER_LENGTH  = 17                            // 默认消息头长度, 不包括 PREFIX_LENGTH
	CONTEXT_OFFSET = PREFIX_LENGTH + HEADER_LENGTH // 自定义上下文内容所在位置，   23
)

func unPackFlag(buffer []byte) (flag uint8) {
	binary.Read(bytes.NewBuffer(buffer[:1]) , binary.BigEndian, &flag)
	return
}

func unPackVer(buffer []byte) (ver uint8) {
	binary.Read(bytes.NewBuffer(buffer[1:2]) , binary.BigEndian, &ver)
	return
}

func unPackWorkId(buffer []byte) (workId uint16) {
	binary.Read(bytes.NewBuffer(buffer[CONTEXT_OFFSET:CONTEXT_OFFSET+2]) , binary.BigEndian, &workId)
	return
}

func unPackId(buffer []byte) (workId uint16) {
	binary.Read(bytes.NewBuffer(buffer[PREFIX_LENGTH+8:PREFIX_LENGTH+12]) , binary.BigEndian, &workId)
	return
}

func NewRequest(buffer []byte) *SRequest {
	req := new(SRequest)
	reader := bytes.NewReader(buffer)
	binary.Read(reader, binary.BigEndian, req)
	return req
}

func (req *SResponse)Pack() []byte {
	var writer = new(bytes.Buffer)
	binary.Write(writer, binary.BigEndian, req)
	return writer.Bytes()
}

func Pack(req interface{}) []byte {
	var writer = new(bytes.Buffer)
	binary.Write(writer, binary.BigEndian, &req)
	return writer.Bytes()
}
