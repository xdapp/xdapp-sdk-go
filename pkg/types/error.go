package types

import (
	"errors"
)

var (
	ErrRequireApp = errors.New("config app is nil")
	ErrRequireServiceName = errors.New("config service name is nil")
	ErrRequireServiceKey = errors.New("config service key is nil")


	ErrReadByteEmpty = errors.New("读取数据为空")
)