package register

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/leesper/tao"
	"github.com/xdapp/xdapp-sdk-go/pkg/types"
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

// MessageNumber returns the message number.
func (req Request) MessageNumber() int32 {
	return int32(config.Version)
}

func (req Request) Serialize() ([]byte, error) {
	var w = new(bytes.Buffer)
	binary.Write(w, binary.BigEndian, req.Prefix)
	binary.Write(w, binary.BigEndian, req.Header)
	return BytesCombine(w.Bytes(), req.Context, req.Body), nil
}

func unSerialize(d []byte) (tao.Message, error) {
	if d == nil {
		return nil, tao.ErrNilData
	}

	req := new(Request)
	reader := bytes.NewReader(d)
	binary.Read(reader, binary.BigEndian, req)
	return req, nil
}

func (req Request) Decode(raw net.Conn) (tao.Message, error) {
	byteChan := make(chan []byte)
	errorChan := make(chan error)

	go func(bc chan []byte, ec chan error) {
		buf, err := readBytesByLength(raw, types.ProtocolPrefixLength)
		if err != nil {
			ec <- err
			close(bc)
			close(ec)
			Logger.Warn("数据解析失败 err: " + err.Error())
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
			return nil, types.ErrReadByteEmpty
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

		headBytes, err := readBytesByLength(raw, types.ProtocolHeaderLength)
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

		body, err := readBytesByLength(raw, int(prefix.Length)-types.ProtocolHeaderLength-ctxLen)
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
func transportRpcRequest(c tao.WriteCloser, flag byte, ver byte, header Header, context []byte, body []byte) error {
	bodyLen := len(body)
	flag = flag | types.ProtocolFlagResultMode

	if bodyLen < types.ProtocolSendChunkLength {
		prefix := Prefix{
			Flag:   uint8(flag | types.ProtocolFlagFinish),
			Ver:    ver,
			Length: uint32(types.ProtocolHeaderLength + len(context) + len(body)),
		}

		err := c.Write(Request{
			Prefix:  prefix,
			Header:  header,
			Context: context,
			Body:    body,
		})
		if err != nil {
			return err
		}
	}

	// 大于 拆包分段发送
	for i := 0; i < bodyLen; i += types.ProtocolSendChunkLength {
		sendLen := Min(types.ProtocolSendChunkLength, bodyLen-i)
		if bodyLen - i == sendLen {
			flag |= types.ProtocolFlagFinish
		}

		chunk := body[i:sendLen]
		prefix := Prefix{
			Flag:   uint8(flag),
			Ver:    ver,
			Length: uint32(types.ProtocolHeaderLength + len(context) + len(chunk)),
		}
		err := c.Write(Request{
			Prefix:  prefix,
			Header:  header,
			Context: context,
			Body:    chunk,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
