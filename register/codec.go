package register

import (
	"net"
	"bytes"
	"errors"
	"encoding/binary"
	"github.com/leesper/tao"
	"github.com/leesper/holmes"
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

type Request struct {
	Prefix
	Header		   // 头信息
	Context []byte // 自定义内容
	Body    []byte // 正文
}

type Prefix struct {
	Flag    byte   // 标志位 成功 |= 4
	Ver     byte   // 版本
	Length  uint32 // 长度
}

type Header struct {
	AppId         uint32
	ServiceId     uint32
	RequestId     uint32
	AdminId       uint32
	ContextLength byte
}

const (
	PREFIX_LENGTH  = 6                             // Flag 1字节、 Ver 1字节、 Length 4字节、HeaderLength 1字节
	HEADER_LENGTH  = 17                            // 默认消息头长度, 不包括 PREFIX_LENGTH
	CONTEXT_OFFSET = PREFIX_LENGTH + HEADER_LENGTH // 自定义上下文内容所在位置，   23
)

// MessageNumber returns the message number.
func (req Request) MessageNumber() int32 {
	return int32(1)
}

// Serialize Request
func (req Request) Serialize() ([]byte, error) {
	var writer = new(bytes.Buffer)
	pack(writer, req.Prefix)
	pack(writer, req.Header)
	return BytesCombine(writer.Bytes(), req.Context, req.Body), nil
}

// Deserialize
func DeserializeRequest(data []byte) (tao.Message, error) {
	if data == nil {
		return nil, tao.ErrNilData
	}

	req := new(Request)
	reader := bytes.NewReader(data)
	binary.Read(reader, binary.BigEndian, req)
	return req, nil
}

// 标识   | 版本    | 长度    | 头信息       | 自定义内容    |  正文
// ------|--------|---------|------------|-------------|-------------
// Flag  | Ver    | Length  | Header     | Context      | Body
// 1     | 1      | 4       | 17         | 默认0，不定   | 不定
// C     | C      | N       |            |             |
//
//
// 其中 Header 部分包括
//
// AppId     | 服务ID      | rpc请求序号  | 管理员ID      | 自定义信息长度
// ----------|------------|------------|-------------|-----------------
// AppId     | ServiceId  | RequestId  | AdminId     | ContextLength
// 4         | 4          | 4          | 4           | 1
// N         | N          | N          | N           | C

func (req Request) Decode(raw net.Conn) (tao.Message, error) {
	byteChan := make(chan []byte)
	errorChan := make(chan error)

	go func(bc chan []byte, ec chan error) {
		buf := make([]byte, tcpConfig.packageMaxLength)
		_, err := raw.Read(buf)
		if err != nil {
			ec <- err
			close(bc)
			close(ec)
			holmes.Debugln("read failed")
			return
		}
		bc <- buf
	}(byteChan, errorChan)

	var readBytes []byte
	select {
	case err := <-errorChan:
		return nil, err

	case readBytes = <-byteChan:
		if readBytes == nil {
			holmes.Warnln("read type bytes nil")
			return nil, errors.New("more than 8M data")
		}

		var prefix Prefix
		err := binary.Read(bytes.NewReader(readBytes[:PREFIX_LENGTH]), binary.BigEndian, &prefix); if err != nil {
			return nil, err
		}
		if prefix.Ver != 1 {
			holmes.Warnln("消息版本错误", prefix.Ver)
			return nil, err
		}

		if prefix.Length > uint32(tcpConfig.packageMaxLength) {
			holmes.Errorf("message(type %d) has bytes(%d) beyond max %d\n", request.Ver, prefix.Length, tcpConfig.packageMaxLength)
			return nil, tao.ErrBadData
		}

		var header Header
		headerBuf := bytes.NewReader(readBytes[PREFIX_LENGTH:CONTEXT_OFFSET])
		if err = binary.Read(headerBuf, binary.BigEndian, &header); err != nil {
			return nil, err
		}
		ctxLen := int(header.ContextLength)

		var request Request
		request.Prefix  = prefix
		request.Header  = header
		request.Context = readBytes[CONTEXT_OFFSET:(CONTEXT_OFFSET+ctxLen)]
		request.Body    = readBytes[(CONTEXT_OFFSET+ctxLen):(PREFIX_LENGTH + prefix.Length)]

		return request, nil
	}
}

// Encode encodes the message into bytes data.
func (req Request) Encode(msg tao.Message) ([]byte, error) {
	data, err := msg.Serialize()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func setRequest(flag byte, ver byte, header Header, context []byte, body[]byte) Request {
	return Request{
		Prefix: Prefix{
			flag,
			ver,
			uint32(HEADER_LENGTH + len(context) + len(body)),
		},
		Header: header, Context: context, Body: body,
	}
}


/**

 */
func socketSend(flag byte, ver byte, header Header, context []byte, data[]byte) {

	flag = flag | FLAG_RESULT_MODE
	dataLength := len(data)

	if dataLength < 0x200000 {
		socketSendChan<-Request{
			Prefix: Prefix{
				Ver: ver,
				Flag: uint8(flag | FLAG_FINISH),
				Length: uint32(HEADER_LENGTH + len(context) + dataLength),
			},
			Header: header, Context: context, Body: data}
		return
	}

	// 大于 拆包分段发送
	for i := 0; i < dataLength; i += 0x200000 {
		chunkLen := Min(0x200000, dataLength-i)
		chunk := data[i:chunkLen]
		if dataLength-i == chunkLen {
			flag |= FLAG_FINISH
		}

		socketSendChan<-Request{
			Prefix: Prefix{
				Ver: ver,
				Flag: uint8(flag),
				Length: uint32(HEADER_LENGTH + len(context) + chunkLen),
			},
			Header: header, Context: context, Body: chunk}
	}
}