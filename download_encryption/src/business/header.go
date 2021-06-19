package business

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_CHECK_TIME = 300 //刷新配置事件（秒）
)

type RequestMessage struct {

	//请求类型
	MessageId int

	//请求的远端信息
	RemoteIp string

	//请求的账号信息
	Account string
	Session string

	//http请求信息
	Writer http.ResponseWriter
}

type RequestService struct {
	Account string `json:"Account"`
	Session string `json:"Session"`
}

const (
	ErrorOK = iota

	ErrorParamNotFound = iota + 1000
	ErrorParamInvalid
	ErrorServerError
	ErrorJsonParseError
	ErrorRedisError
	ErrorMysqlError
	ErrorAppIdNotFound
	ErrorSignatureCheckFail
	ErrorFileExist
	ErrorBokerchainNotInitialized
	ErrorContractCallError
	ErrorAccountExist
	ErrorAccountNotFound
	ErrorAccountBinded
	ErrorNotAuthorized
	ErrorAppNotFound
	ErrorShaNotMatch
	ErrorNotMatch
	ErrorOrderIdExist
)

type ErrorContext struct {
	Code    int
	Message string
}
type Error *ErrorContext

func ByteToFixedByte(src []byte) (dst [32]byte) {

	for i := 0; i < 32; i++ {

		if i < len(src) {
			dst[i] = src[i]
		} else {
			dst[i] = 0
		}
	}
	return dst
}

// execute cmd line
func ShellExecute(s string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", s)

	var cout bytes.Buffer
	cmd.Stdout = &cout

	var cerr bytes.Buffer
	cmd.Stderr = &cerr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return cout.String(), nil
}

func HextoByte(str string) []byte {

	slen := len(str)
	bHex := make([]byte, len(str)/2)
	ii := 0
	for i := 0; i < len(str); i = i + 2 {
		if slen != 1 {
			ss := string(str[i]) + string(str[i+1])
			bt, _ := strconv.ParseInt(ss, 16, 32)
			bHex[ii] = byte(bt)
			ii = ii + 1
			slen = slen - 2
		}
	}
	return bHex
}

func NewError(code int, format string, args ...interface{}) Error {
	return &ErrorContext{code, fmt.Sprintf(format, args...)}
}

func HttpErrorEx(w http.ResponseWriter, err Error) {
	w.Write([]byte(fmt.Sprintf("{\"result\":%d,\"msg\":\"%s\"}", err.Code, err.Message)))
}

func HttpError(w http.ResponseWriter, result int, msg string) {
	w.Write([]byte(fmt.Sprintf("{\"result\":%d,\"msg\":\"%s\"}", result, msg)))
}

func HttpFormGetString(r *http.Request, param string) (value string, err Error) {
	if len(r.Form[param]) <= 0 {
		return "", NewError(ErrorParamNotFound, "param %s not found", param)
	}
	return strings.TrimSpace(r.Form[param][0]), nil
}

func HttpFormGetInt(r *http.Request, param string) (value int, err Error) {
	if len(r.Form[param]) <= 0 {
		return 0, NewError(ErrorParamNotFound, "param %s not found", param)
	}
	valueStr := strings.TrimSpace(r.Form[param][0])
	value, e := strconv.Atoi(valueStr)
	if e != nil {
		return 0, NewError(ErrorParamInvalid, "param %s invalid err=%s", param, e.Error())
	}
	return value, nil
}

func HttpFormGetInt64(r *http.Request, param string) (value int64, err Error) {
	if len(r.Form[param]) <= 0 {
		return 0, NewError(ErrorParamNotFound, "param %s not found", param)
	}
	valueStr := strings.TrimSpace(r.Form[param][0])
	value, e := strconv.ParseInt(valueStr, 10, 64)
	if e != nil {
		return 0, NewError(ErrorParamInvalid, "param %s invalid err=%s", param, e.Error())
	}
	return value, nil
}

func PathBase(filePath string) string {
	filePath = strings.TrimRight(filePath, "/\\")
	if filePath == "" {
		return "."
	}

	idx1 := strings.LastIndex(filePath, "/")
	idx2 := strings.LastIndex(filePath, "\\")
	idx := idx1
	if idx2 > idx {
		idx = idx2
	}

	if idx >= 0 {
		filePath = filePath[idx+1:]
	}
	if filePath == "" {
		return "/"
	}
	return filePath
}

func GetExeName() string {
	ret := ""
	ex, err := os.Executable()
	if err == nil {
		ret = filepath.Base(ex)
	}
	return ret
}

func TimeToString(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func TimestampToString(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}
func TimeFromString(str string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
}

func NowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func PageRange(total, page, pageSize int) (start, end int) {
	if total <= 0 {
		return 0, 0
	}

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		start = 0
		end = total - 1
		return start, end
	}

	start = (page - 1) * pageSize
	if start >= total {
		page := total/pageSize + 1
		if 0 == total%pageSize {
			page = page - 1
		}
		start = (page - 1) * pageSize
	}

	end = start + pageSize - 1
	if end >= total {
		end = total - 1
	}

	return start, end
}

func PathExist(fullPath string) bool {
	_, err := os.Stat(fullPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

/*
// HasLocalIPddr 检测 IP 地址字符串是否是内网地址
func HasLocalIPddr(ip string) bool {
	return HasLocalIP(net.ParseIP(ip))
}

// HasLocalIP 检测 IP 地址是否是内网地址
func HasLocalIP(ip net.IP) bool {
	for _, network := range localNetworks {
		if network.Contains(ip) {
			return true
		}
	}

	return ip.IsLoopback()
}

// ClientIP 尽最大努力实现获取客户端 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// ClientPublicIP 尽最大努力实现获取客户端公网 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" && !HasLocalIPddr(ip) {
			return ip
		}
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" && !HasLocalIPddr(ip) {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		if !HasLocalIPddr(ip) {
			return ip
		}
	}

	return ""
}*/

// RemoteIP 通过 RemoteAddr 获取 IP 地址， 只是一个快速解析方法。
func RemoteIP(r *http.Request) string {
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}
