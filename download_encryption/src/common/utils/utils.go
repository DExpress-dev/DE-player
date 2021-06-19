package utils

import (
	"bytes"
	"common/config/goini"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "common/log4go"

	//log4plus "github.com/alecthomas/log4go"
	"github.com/axgle/mahonia"
)

//const define
const (
	VOD              = 0
	LIVE             = 1
	ERR_HTTP_URL     = 2
	ERR_HTTP_CONTENT = 3
	ERR_RET          = 10
)

//var Log log4plus.Logger
var ConfigFile *goini.Config

func init() {
	ConfigFile = goini.Init("config.ini")

	/*logFileName := ConfigFile.Read_string("log", "name", "go.log")

	Log = log4plus.NewLogger()

	err := os.MkdirAll("/var/log/go", os.ModePerm)
	if err != nil {
		fmt.Println("dealCompress MkdirAll, err:%v", err)
	}

	fileWriter := log4plus.NewFileLogWriter("/var/log/go/"+logFileName, true)
	consoleWriter := log4plus.NewConsoleLogWriter()
	fileWriter.SetRotateDaily(true)
	Log.AddFilter("file", log4plus.DEBUG, fileWriter)       //输出到file,级别为DEBUG
	Log.AddFilter("console", log4plus.DEBUG, consoleWriter) //输出到console,级别为DEBUG
        */

}

//获取一个随机数
func GetRandNum() string {
	randNum := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strconv.Itoa(randNum.Intn(1000000))
}

//清零空的字符
func ByteString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

//清理空格 回车 换行
func ClearSpecialChar(str string) string {
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\r", "", -1)

	return str
}

// 返回true，路径存在
// 返回false并且无错，路径不存在
// 返回错误，不确定路径是否存在
func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//从其他服务器拿到的文件，临时存放目录
var smallFilesPath string = "/data/smallFilesDir/"

func GenerateSmallFileName(inputUrl string) string {
	inputUrl = ClearSpecialChar(inputUrl)
	urlParsed, err := url.Parse(inputUrl)
	if err != nil {
		fmt.Println(" url.Parse:", inputUrl, "err")
		return ""
	}

	return smallFilesPath + urlParsed.Path
}

func GetSmallFilePath() string {
	return smallFilesPath
}

//转码后的文件存放路径
var dstFilePath string = "/data/otv/dstFileDir/"

func GetDstFilePath() string {
	return dstFilePath
}

func DealPanic() {
	var err error
	r := recover()
	if r != nil {
		switch t := r.(type) {
		case string:
			err = errors.New(t)
		case error:
			err = t
		default:
			err = errors.New("Unknown error")
		}

		fmt.Println("in go process panic:", err.Error())
	}
}

// 获得目录路径。若路径为空，错误；若不含/，返回当前路径'.'
// 例：若path为"/a/b/c/d/index.m3u8////"，dirName为"/a/b/c/d"
func GetDirName(path string) (dirName string, err error) {
	i := len(path) - 1
	if i < 0 {
		return "", errors.New("path is nil")
	}
	// Remove trailing slashes
	for ; i > 0 && path[i] == '/'; i-- {
		path = path[:i]
	}

	i = strings.LastIndex(path, "/")

	if i < 0 {
		return ".", nil
	}

	dirName = path[:i]
	return dirName, nil
}

// 对外暴露的接口：
// 将http请求中的包体转换为json格式，但是不清空包体内容，还可以重新使用包体
func ParseReqBodyToJsonUnclosed(r *http.Request, bodyStruct interface{}, delBody bool) bool {

	if r.ContentLength <= 0 {
		return false
	}
	var bodySlc []byte = make([]byte, 1024)
	bodyLen, _ := r.Body.Read(bodySlc)
	bodySlc = bodySlc[:bodyLen]
	str := string(bodySlc)

	err := json.Unmarshal([]byte(str), bodyStruct)
	if err == nil {
		return true
	} else {
		return false
	}
}

// 对外暴露的接口：
// 将http请求中的包体转换为json格式，解决了包体中含有中文字符的情况
func ParseReqBodyToJson(r *http.Request, bodyStruct interface{}, delBody bool) bool {
	if delBody {
		defer r.Body.Close()
	}

	if r.ContentLength <= 0 {
		return false
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return false
	}

	//fmt.Println(string(body))
	return ParseBodyToJson(body, bodyStruct)
}

// 背景：在 Golang 的调试过程中出现中文乱码
// 原因：Golang 默认不支持 UTF-8 以外的字符集
// 解决：将字符串的编码转换成UTF-8，使用第三方库github.com/axgle/mahonia
// src 字符串
// srcCode 先对字符串按srcCode格式解码
// tagCode 将上面解码的结果，再转换为tagCode格式
func convertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

//http get function to return response body string
func HttpGet(url string) string {

	client := http.Client{
		//add 5s timeout
		Timeout: time.Duration(5 * time.Second),
	}
	res, err := client.Get(url)
	//res, err := http.Get(url)
	if err != nil {
		log.Error("in httpGet,error:%s\n", err.Error())
		return ""
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error("http read body error:", err.Error())
		return ""
	}
	log.Debug("httpGet url:%s,resp:%s", url, string(body))
	return string(body)
}

//解析 body[]byte 到json格式
func ParseBodyToJson(body []byte, bodyStruct interface{}) bool {
	str := ByteString(body)
	str = convertToString(str, "gbk", "utf-8")

	err := json.Unmarshal([]byte(str), &bodyStruct)
	if err != nil {
		log.Error("unmarshal str:%s failed!", str)
		return false
	}

	return true
}

// http 上传文件服务端实现
// 返回参数，上传文件在本地的绝对路径
func UploadReceive(r *http.Request, saveDir string) (string, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		fmt.Println("ParseMultipartForm", err)
		return "", err
	}
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return "", err
	}

	path := saveDir + "/" + handler.Filename
	fd, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer fd.Close()

	_, err = io.Copy(fd, file)
	if err != nil {
		return "", err
	}

	return path, nil
}

// http 上传文件客户端实现
func UploadFile(filePath string, targetUrl string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	filename := filepath.Base(filePath)
	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer fh.Close()

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

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

//把检索到数据表的内容转化成为JSON格式
func RowsToJson(rows *sql.Rows) string {
	columns, err := rows.Columns()
	if err != nil {
		return ""
	}

	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}

	var jsonData []byte
	if jsonData, err = json.Marshal(tableData); err != nil {
		fmt.Println("2 row to json:", err)
		return ""
	}

	return string(jsonData)
}

//echo -n "value" | openssl sha1 -hmac "key"
//hmac_sha1 此函数名是对应外在的函数名,作用类似md5
func HmacSha1(content, key string) []byte {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(content))

	return mac.Sum(nil)
}

// inputUrl: http://192.168.1.1/otv/aba/c/test.mp4
// savePath: /data, if savePath not exist, make it
// absPath: /data/test.mp4
func DownloadFile(inputUrl, savePath string) (absPath string, err error) {
	statusCode := ExeCmdAndStdOut("curl -I -m 3 -s -o /dev/null -w %{http_code} " + inputUrl)

	if statusCode != "200" {
		return "", errors.New("Connection timed out after 3001 milliseconds")
	}

	savePath = strings.TrimSuffix(savePath, "/")

	if b, _ := PathExist(savePath); !b {
		err = os.MkdirAll(savePath, 0755)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
	}

	baseName := filepath.Base(inputUrl)
	absPath = savePath + "/" + baseName

	var strCmd string
	if strings.HasPrefix(inputUrl, "http://") {
		strCmd = fmt.Sprintf("wget -q -c -t 0 -O %s %s", absPath, inputUrl)

	} else if strings.HasPrefix(inputUrl, "ftp://") {
		strCmd = fmt.Sprintf("wget -q -c -t 0 -O %s %s", absPath, inputUrl)

	} else if strings.HasPrefix(inputUrl, "https://") {
		strCmd = fmt.Sprintf("curl -L -s -o %s %s", absPath, inputUrl)

	} else {
		return "", errors.New("input url not http/https/ftp")
	}

	if b := ExeCmd(strCmd); !b {
		return "", errors.New("download fail " + strCmd)
	}

	return absPath, nil
}

//  deprecated!!!  下载文件，返回保存路径
func DoLoadFile(inputUrl string) string {
	statusCode := ExeCmdAndStdOut("curl -I -m 3 -s -o /dev/null -w %{http_code} " + inputUrl)

	if statusCode != "200" {
		return ""
	}

	pos := strings.Index(inputUrl, " ")
	if pos != -1 {
		inputUrl = ClearSpecialChar(inputUrl)
	}

	path := GenerateSmallFileName(inputUrl)
	if path == "" {
		return ""
	}

	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	var strCmd string

	if strings.HasPrefix(inputUrl, "http://") {
		//strCmd = fmt.Sprintf("curl -L -s --limit-rate %d --retry 30 --retry-delay 1 -o %s %s", speed, path, inputUrl)
		strCmd = fmt.Sprintf("wget -q -c -t 0 -O %s %s", path, inputUrl)

	} else if strings.HasPrefix(inputUrl, "ftp://") {

		//strCmd = fmt.Sprintf("curl %s -u %s:%s -o %s", inputUrl, name, password, path)
		strCmd = fmt.Sprintf("wget -q -c -t 0 -O %s %s", path, inputUrl)
	} else if strings.HasPrefix(inputUrl, "https://") {
		strCmd = fmt.Sprintf("curl -L -s -o %s %s", path, inputUrl)
	} else {
		return ""
	}

	b := ExeCmd(strCmd)
	if !b {
		return ""
	}

	return path
}

/*symbollink director */
func SoftLink(src string, link string) error {
	b, _ := PathExist(link)
	if !b {
		err := os.Symlink(src, link)
		return err
	}
	return nil
}

/*
create directory
dir	directory
ret	true :suc ,false: faild*/
func CreateDir(dir string) error {
	b, _ := PathExist(dir)
	if !b {
		err := os.MkdirAll(dir, os.ModePerm)
		return err
	}
	return nil
}

func checkStatus(url string) int {
	resp, err := http.Head(url)
	if err != nil {
		fmt.Println("Error:", err)
		return -1
	}
	return resp.StatusCode
}

/*check whether live m3u8 . check the wether have #EXT-X-ENDLIST
ret:
0:vod  1:live
*/
func CheckLive(url string) int {
	httpStatus := checkStatus(url)
	if httpStatus != 200 {
		return httpStatus
	}
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return ERR_HTTP_URL
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if strings.Contains(string(body), "PROGRAM") {
		secM3u8 := getSecondM3u8Url(url)
		if secM3u8 == "" {
			return ERR_HTTP_URL
		}
		return CheckLive(secM3u8)
	}
	if strings.Contains(string(body), "#EXT-X-ENDLIST") {
		return VOD
	} else {
		return LIVE
	}
}

/*
retun second m3u8 if orgin is  first m3u8 ,or return self if orgin is second m3u8
*/
func getSecondM3u8Url(oriUrl string) string {
	baseUrl, err := GetDirName(oriUrl)
	if err != nil {
		fmt.Println("getSecondM3u8Url", err)
	}
	resp, err := http.Get(oriUrl)
	if err != nil {
		fmt.Println("getSecondM3u8Url", err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("getSecondM3u8Url, read err", err)
	}
	lines := strings.Split(string(content), "\n")
	var i int
	var topM3u8 bool
	for i = len(lines) - 1; i >= 0; i-- {
		if strings.HasPrefix(lines[i], "#EXT-X-STREAM-INF") {
			topM3u8 = true
			btrLine := strings.TrimSpace(lines[i+1])
			secUrl := baseUrl + "/" + btrLine
			resp, err := http.Get(secUrl)
			if err != nil {
				fmt.Println(err)
				continue
			}
			defer resp.Body.Close()
			resCode := resp.StatusCode
			if resCode != 200 {
				continue
			} else {
				return secUrl
			}
		}
	}
	if topM3u8 && i == -1 {
		return ""
	}
	return oriUrl
}

func GetHttpProxyConfig() (string, error) {
	return os.Getenv("http_proxy"), nil
}

func GetHttpProxy() func(*http.Request) (*url.URL, error) {
	proxy, err := GetHttpProxyConfig()
	if err != nil {
		return nil
	}
	proxy = strings.TrimSpace(proxy)
	if "" == proxy {
		return nil
	}
	return func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxy)
	}
}

//分行
func SplitLine(str string) []string {
	return strings.FieldsFunc(str, func(s rune) bool {
		if s == '\n' || s == '\r' {
			return true
		}
		return false
	})
}

func MD5(b []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(b)
	return hex.EncodeToString(md5Ctx.Sum(nil))
}

func NowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func FileSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}
