package main

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//判断文件是否存在
func FileExist(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			return false
		}
	}
	return true
}

//获取当前路径
func GetCurrentDirectory() string {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}

	return strings.Replace(dir, "\\", "/", -1)
}

func SaveFile(content, localFile string) error {
	localPath := path.Dir(localFile)
	err := os.MkdirAll(localPath, os.ModePerm)
	if err != nil {
		return err
	}

	localFileTmp := localFile + ".tmp"
	file, err := os.Create(localFileTmp)
	if err != nil {
		return err
	}
	io.WriteString(file, content)
	file.Close()
	os.Rename(localFileTmp, localFile)
	return nil
}
