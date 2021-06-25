package httputils

import (
	"bufio"
	"bytes"
	"common/utils"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func HttpRequestForm(r *http.Request, param string) (value string, err error) {
	if len(r.Form[param]) <= 0 {
		return "", fmt.Errorf("param %s not found!", param)
	}
	return strings.TrimSpace(r.Form[param][0]), nil
}

func PathOfUrl(remoteUrl string) string {
	urlInfo, err := url.Parse(remoteUrl)
	if err != nil {
		return ""
	}
	return urlInfo.Path
}

func ClientAddress(req *http.Request) string {
	ip := req.Header.Get("X-Real-IP")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = req.Header.Get("Remote_addr")
			if ip == "" {
				ip = strings.Split(req.RemoteAddr, ":")[0]
			}
		}
	}
	return ip
}

func httpClient(timeout int) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:           utils.GetHttpProxy(),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Dial: func(netw, addr string) (net.Conn, error) {
				to := time.Duration(timeout) * time.Second
				conn, err := net.DialTimeout(netw, addr, to)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(to))
				return conn, nil
			},
		},
	}
}

func httpGet(remote string, timeout int, headers http.Header, resBody io.Writer) (size int64, err error) {

	request, err := http.NewRequest("GET", remote, nil)
	if err != nil {
		return 0, err
	}
	request.Close = true
	request.Header.Set("Connection", "close") // 完成后断开连接
	if headers != nil {
		for key, header := range headers {
			request.Header.Set(key, header[0])
		}
	}

	response, err := httpClient(timeout).Do(request)
	if err != nil {
		return 0, err
	}
	// 保证I/O正常关闭
	defer response.Body.Close()
	size, err = io.Copy(resBody, response.Body)
	if err != nil {
		return 0, err
	}

	if http.StatusOK != response.StatusCode {
		err = fmt.Errorf("%s", response.Status)
	}

	return size, err
}

func HttpGet(remote string, timeout int, headers http.Header) (string, error) {
	buf := new(bytes.Buffer)
	_, err := httpGet(remote, timeout, headers, buf)
	if err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}

// download to buffer
// remote 远端文件路径
// timeout 下载超时
// buf
func DownloadBuffer(remote string, timeout int, buf *bytes.Buffer) (int64, error) {
	written, err := httpGet(remote, timeout, nil, buf)
	if err != nil {
		return 0, fmt.Errorf("DownloadBuffer : %s", err.Error())
	}

	return written, nil
}

//加密过程：
//  1、处理数据，对数据进行填充，采用PKCS7（当密钥长度不够时，缺几位补几个几）的方式。
//  2、对数据进行加密，采用AES加密方法中CBC加密模式
//  3、对得到的加密数据，进行base64加密，得到字符串
// 解密过程相反

//16,24,32位字符串的话，分别对应AES-128，AES-192，AES-256 加密方法
//key不能泄露
//var PwdKey = []byte("ABCDABCDABCDABCD")

//pkcs7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	//判断缺少几位长度。最少1，最多 blockSize
	padding := blockSize - len(data)%blockSize
	//补足位数。把切片[]byte{byte(padding)}复制padding个
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

//pkcs7UnPadding 填充的反向操作
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	}
	//获取填充的个数
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}

//AesEncrypt 加密
func AesEncrypt(data []byte, key []byte) ([]byte, error) {
	//创建加密实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//判断加密快的大小
	blockSize := block.BlockSize()
	//填充
	encryptBytes := pkcs7Padding(data, blockSize)
	//初始化加密数据接收切片
	crypted := make([]byte, len(encryptBytes))
	//使用cbc加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	//执行加密
	blockMode.CryptBlocks(crypted, encryptBytes)
	return crypted, nil
}

//AesDecrypt 解密
func AesDecrypt(data []byte, key []byte) ([]byte, error) {
	//创建实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块的大小
	blockSize := block.BlockSize()
	//使用cbc
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	//初始化解密数据接收切片
	crypted := make([]byte, len(data))
	//执行解密
	blockMode.CryptBlocks(crypted, data)
	//去除填充
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}

//EncryptByAes Aes加密 后 base64 再加
func EncryptByAes(data []byte, key string) (string, error) {
	res, err := AesEncrypt(data, []byte(key))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(res), nil
}

//DecryptByAes Aes 解密
func DecryptByAes(data string, key string) ([]byte, error) {
	dataByte, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return AesDecrypt(dataByte, []byte(key))
}

func DecryptFile(srcFile, key, destFile string) error {

	var err error
	encryptionPath := path.Dir(destFile)
	if err = os.MkdirAll(encryptionPath, os.ModePerm); err != nil {
		return err
	}

	var f *os.File
	if f, err = os.Open(srcFile); err != nil {
		return err
	}
	defer f.Close()

	//	fInfo, _ := f.Stat()
	br := bufio.NewReader(f)

	var ff *os.File
	if ff, err = os.Create(destFile); err != nil {
		return err
	}
	defer ff.Close()

	num := 0
	for {
		num = num + 1
		a, err := br.ReadString('\n')
		if err != nil {
			break
		}
		getByte, err := DecryptByAes(a, key)
		if err != nil {
			return err
		}

		buf := bufio.NewWriter(ff)
		buf.Write(getByte)
		buf.Flush()
	}
	//	ffInfo, _ := ff.Stat()
	return nil
}

func EncryptionFile(srcFile, key, destFile string) error {

	var err error
	encryptionPath := path.Dir(destFile)
	if err = os.MkdirAll(encryptionPath, os.ModePerm); err != nil {
		return err
	}
	content, err := os.Open(srcFile)

	maxLen := 1024 * 1024 * 100
	srcFileInfo, _ := content.Stat()
	fLen := srcFileInfo.Size()
	var forCount int64 = 0
	getLen := fLen
	if fLen > int64(maxLen) {
		getLen = int64(maxLen)
		forCount = fLen / int64(maxLen)
	}

	var file *os.File
	if file, err = os.Create(destFile); err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < int(forCount+1); i++ {
		srcBuffer := make([]byte, getLen)
		n, err := content.Read(srcBuffer)
		if err != nil {
			return err
		}
		getByte, err := EncryptByAes(srcBuffer[:n], key)
		if err != nil {
			fmt.Printf("error=%s \n", err.Error())
			return err
		}
		getBytes := append([]byte(getByte), []byte("\n")...)

		buf := bufio.NewWriter(file)
		buf.WriteString(string(getBytes[:]))
		buf.Flush()
	}
	return nil
}

// download to file
// localFile 本地保存路径
// remote 远端文件路径
// timeout 下载超时
func DownloadFile(remote, srcFile, key, encryptionFile string, timeout int) (int64, error) {

	srcPath := path.Dir(srcFile)
	err := os.MkdirAll(srcPath, os.ModePerm)
	if err != nil {
		return 0, err
	}

	file, err := os.Create(srcFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	//将文件信息写入相关文件中
	written, err := httpGet(remote, timeout, nil, file)
	if err != nil {
		return 0, fmt.Errorf("DownloadFile : %s", err.Error())
	}

	//判断进行加密
	if key != "" {
		EncryptionFile(srcFile, key, encryptionFile)
		//		DecryptFile(encryptionFile, key, srcFile+"_222")
	}

	return written, nil
}
