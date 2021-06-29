package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

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
func AesEncrypt(data []byte, key string, iv string) ([]byte, error) {

	//创建加密实例
	block, err := aes.NewCipher([]byte(key))
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
	blockMode := cipher.NewCBCEncrypter(block, []byte(iv))
	//执行加密
	blockMode.CryptBlocks(crypted, encryptBytes)

	return crypted, nil
}

//AesDecrypt 解密
func AesDecrypt(data []byte, key string, iv string) ([]byte, error) {
	//创建实例
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	//使用cbc
	blockMode := cipher.NewCBCDecrypter(block, []byte(iv))
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
func EncryptByAes(data []byte, key string, iv string) (string, error) {

	res, err := AesEncrypt(data, key, iv)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(res), nil
}

//DecryptByAes Aes 解密
func DecryptByAes(data string, key string, iv string) ([]byte, error) {

	dataByte, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return AesDecrypt(dataByte, key, iv)
}
