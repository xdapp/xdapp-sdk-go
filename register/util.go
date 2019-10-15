package register

import (
	"bytes"
	"encoding/binary"
	"io"
	"strconv"
)

func IntToStr(data interface{}) string {
	switch value := data.(type) {
	case int:
		return strconv.Itoa(value) // int to str
	case int64:
		return strconv.FormatInt(value, 10) // int64 to str
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	default:
		return ""
	}
}

func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// BytesCombine 多个[]byte数组合并成一个[]byte
func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func pack(w io.Writer, data interface{}) error {
	return binary.Write(w, binary.BigEndian, data)
}