package register

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/leesper/tao"
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
	Header         // 头信息
	Context []byte // 自定义内容
	Body    []byte // 正文
}

type Prefix struct {
	Flag   byte   // 标志位 成功 |= 4
	Ver    byte   // 版本
	Length uint32 // 长度
}

type Header struct {
	AppId         uint32
	ServiceId     uint32
	RequestId     uint32
	AdminId       uint32
	ContextLength byte
}

var (
	ErrReadByteEmpty = errors.New("读取数据为空")
)

// MessageNumber returns the message number.
func (req Request) MessageNumber() int32 {
	return int32(config.Version)
}

func (req Request) Serialize() ([]byte, error) {
	var writer = new(bytes.Buffer)
	binary.Write(writer, binary.BigEndian, req.Prefix)
	binary.Write(writer, binary.BigEndian, req.Header)
	return BytesCombine(writer.Bytes(), req.Context, req.Body), nil
}

func Unserialize(data []byte) (tao.Message, error) {
	if data == nil {
		return nil, tao.ErrNilData
	}

	req := new(Request)
	reader := bytes.NewReader(data)
	binary.Read(reader, binary.BigEndian, req)
	return req, nil
}

func (req Request) Decode(raw net.Conn) (tao.Message, error) {
	byteChan := make(chan []byte)
	errorChan := make(chan error)

	go func(bc chan []byte, ec chan error) {
		buf, err := readBytesByLength(raw, PrefixLength)
		if err != nil {
			ec <- err
			close(bc)
			close(ec)
			Logger.Warn("read failed")
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
			return nil, ErrReadByteEmpty
		}

		var header Header
		var prefix Prefix

		err := binary.Read(bytes.NewReader(readBytes), binary.BigEndian, &prefix)
		if err != nil {
			return nil, err
		}
		if prefix.Ver != byte(config.Version) {
			return nil, errors.New("消息版本错误" + string(prefix.Ver))
		}

		if prefix.Length > uint32(config.PackageMaxLength) {
			err := fmt.Sprintf("数据长度为%d, 大于最大值%d", prefix.Length, config.PackageMaxLength)
			Logger.Error(err)
			return nil, errors.New(err)
		}

		headBytes, err := readBytesByLength(raw, HeaderLength)
		if err != nil {
			return nil, err
		}
		headBuf := bytes.NewReader(headBytes)
		if err = binary.Read(headBuf, binary.BigEndian, &header); err != nil {
			return nil, err
		}

		ctxLen := int(header.ContextLength)
		context, err := readBytesByLength(raw, ctxLen)
		if err != nil {
			return nil, err
		}

		body, err := readBytesByLength(raw, int(prefix.Length)-HeaderLength-ctxLen)
		if err != nil {
			return nil, err
		}

		return Request{prefix, header, context, body}, nil
	}
}

func readBytesByLength(r io.Reader, len int) ([]byte, error) {
	byte := make([]byte, len)
	_, err := io.ReadFull(r, byte)
	return byte, err
}

// Encode encodes the message into bytes data.
func (req Request) Encode(msg tao.Message) ([]byte, error) {
	data, err := msg.Serialize()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// 转发消息到其它服务
func transportRpcRequest(c tao.WriteCloser, flag byte, ver byte, header Header, context []byte, body []byte) {

	totalLen := len(body)
	flag = flag | FlagResultMode

	if totalLen < SendChunkLength {
		prefix := Prefix{uint8(flag | FlagFinish), ver, uint32(HeaderLength + len(context) + len(body))}
		sendRequest(c, Request{prefix, header, context, body})
		return
	}

	// 大于 拆包分段发送
	for i := 0; i < totalLen; i += SendChunkLength {
		sendLen := Min(SendChunkLength, totalLen-i)
		if totalLen-i == sendLen {
			flag |= FlagFinish
		}

		chunk := body[i:sendLen]
		prefix := Prefix{uint8(flag), ver, uint32(HeaderLength + len(context) + len(chunk))}
		sendRequest(c, Request{prefix, header, context, chunk})
	}
}

func sendRequest(c tao.WriteCloser, request Request) {
	if err := c.Write(request); err != nil {
		Logger.Error("error", err)
	}
}
