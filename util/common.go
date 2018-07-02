package util

import (
	"crypto/sha1"
	"encoding/hex"
	"runtime"
	"path/filepath"
	"strings"
	"encoding/json"
	"fmt"
	"os"
	"crypto/md5"
	"io"
	"time"
	"strconv"
	"path"
)

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

/**
	当前时间
 */
func Time() int64 {
	return time.Now().Unix()
}

/**
	时间字符串形式
 */
func TimeStr() string {
	return IntToStr(Time())
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
	获取当前文件夹下全部文件
 */
func FindAllFiles(dir string) []string {

	pattern := strings.Replace(dir, "\\", "/", -1) + "/*"
	files, _ := filepath.Glob(pattern)
	fmt.Println("FindAllFiles: ", files) // contains a list of all files in the current directory
	return files
}

/**
	获取文件信息  文件名+后缀、后缀
 */
func GetFileInfo(f string) (string, string) {
	return path.Base(f), path.Ext(f)
}