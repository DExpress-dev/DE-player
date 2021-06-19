// stream_check project public.go
package public

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

func CopyFile(src, dest string) (err error) {
	var srcFile *os.File
	srcFile, err = os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	localPath := path.Dir(dest)
	err = os.MkdirAll(localPath, os.ModePerm)
	if err != nil {
		return err
	}

	var dstFile *os.File
	dstFile, err = os.Create(dest)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)

	return err
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
