package register

import (
	"strings"
	"encoding/json"
	"fmt"
	"os"
	"crypto/md5"
	"io"
	"time"
	"strconv"
)

/**

 */
func JsonEncode(data interface{}) string {
	json, err := json.Marshal(data)

	if err != nil {
		fmt.Println(err.Error())
	}

	return string(json)
}

/**

 */
func JsonDecode(str string, fields interface{}) {

	err := json.Unmarshal([]byte(str), &fields)

	if err != nil {
		fmt.Println(err.Error())
	}
}
/**

 */
func Implode(split string, array map[string]string) string {
	var str string
	for _, v := range array {
		str += v + split
	}
	return strings.Trim(str, split)
}

/**
	文件md5
 */
func Md5File(path string) string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	hash := md5.New()
	io.Copy(hash, file)

	return fmt.Sprintf("%x", hash.Sum(nil))
}

/**
	md5
 */
func Md5(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))

	return fmt.Sprintf("%x", hash.Sum(nil))
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

 */
func StrToInt(str string) int {
	data, _ := strconv.Atoi(str)
	return data
}

/**

 */
func StrToInt64(str string) int64 {
	data, _ := strconv.ParseInt(str, 10, 64)
	return data
}

/**
	判断文件存在
 */
func PathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

/**

 */
func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

/**
	字符串截取
 */
func Substr(s string, pos int, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}