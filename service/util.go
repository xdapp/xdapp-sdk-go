package service

import (
	"runtime"
	"path/filepath"
	"strings"
	"crypto/sha1"
	"encoding/hex"
	"time"
	"strconv"
	"fmt"
)

/**
	获取函数名
 */
func GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	funcName = filepath.Ext(funcName)
	funcName = strings.TrimPrefix(funcName, ".")
	return funcName;
}


/**
	获取sha1加密
  */
func Sha1(str string) string {
	getHash := sha1.New()
	getHash.Write([]byte(str))
	r := getHash.Sum(nil)
	return hex.EncodeToString(r[:])
}

/**
	当前时间
 */
func Time() int64 {
	return time.Now().Unix()
}

func IntToStr(data interface{}) string {

	switch value := data.(type) {
	case int:
		return strconv.Itoa(value) // int to str
	case int64:
		return strconv.FormatInt(value, 10) // int64 转str
	default:
		return ""
	}
}

/**
	hash 值
 */
func getHash(app string, name string, time string, rand string, key string) string {
	str := fmt.Sprintf("%s.%s.%s.%s.%s.xdapp.com", app, name, time, rand, key)
	return Sha1(str)
}