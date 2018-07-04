package register

import (
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
	"net/http"
	"io/ioutil"
	"bytes"
	"mime/multipart"
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

/**
	获取当前文件夹下全部文件
 */
func FindAllFiles(dir string) []string {

	pattern := strings.Replace(dir, "\\", "/", -1) + "/*"
	files, _ := filepath.Glob(pattern)
	//MyLog.Info("FindAllFiles: ", files) // contains a list of all files in the current directory
	return files
}

/**
	获取文件信息  文件名+后缀、后缀
 */
func GetFileInfo(f string) (string, string) {
	return path.Base(f), path.Ext(f)
}


/**
	执行curl
 */
func Request(reqUrl string, postStr string) string {
	timeout := time.Duration(5 * time.Second)	//超时时间5s
	client 	:= &http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("POST", reqUrl, strings.NewReader(postStr))
	if err != nil {
		MyLog.Error("执行curl 报错" + err.Error())
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Cookie", "")
	response, err := client.Do(request)
	if err != nil {
		MyLog.Error("执行curl 报错" + err.Error())
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		MyLog.Error("执行curl 读取返回结果报错" + err.Error())
	}

	return string(body)
}

/**
	上传文件
 */
func PostFile(filename string, targetUrl string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		MyLog.Error("执行上传文件  CreateFormFile" + filename + ", 报错" + err.Error())
		return err
	}

	//打开文件句柄操作
	fh, err := os.Open(filename)
	if err != nil {
		MyLog.Error("执行上传文件 打开" + filename + ", 报错" + err.Error())
		return err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	MyLog.Debug("上传文件返回：status=" + resp.Status + ",结果" + string(body))
	return nil
}