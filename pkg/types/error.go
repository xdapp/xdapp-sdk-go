package types

import (
	"errors"
)

var (
	DebugRetryTCPMessage = "[tcp] RPC服务连接关闭，等待重新连接"

	ErrRequireApp         = errors.New("[初始化] 配置app为空")
	ErrRequireServiceName = errors.New("[初始化] 配置service name为空")
	ErrRequireServiceKey  = errors.New("[初始化] 配置service key为空")

	ErrTCPRequireHost  = errors.New("[tcp] 缺少host")
	ErrTCPParseRequest = errors.New("[tcp] 解析数据格式异常")

	ErrReadByteEmpty   = errors.New("[tcp数据解析] 读取数据为空")
	ErrVersionIllegal  = errors.New("[tcp数据解析] 版本异常")
	ErrDataRuleIllegal = errors.New("[tcp数据解析] 数据格式异常")
)