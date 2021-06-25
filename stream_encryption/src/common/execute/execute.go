package execute

import (
	log "common/log4go"
	"common/utils"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//执行系统命令字符串
//字符串中，不可以有连续2个以上的空格
func ExeCmdAndStdOut(strCmd string) string {
	args := strings.Split(strCmd, " ")
	cmd := exec.Command(args[0], args[1:]...)

	fmt.Println(strCmd)

	stdoutPip, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("StdoutPipe Error: %s\n", err)
		log.Error("StdoutPipe :%v, cmd:%s", err, strCmd)
		return ""
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		log.Error("exeCmd:%v", err)
		return ""
	}

	if err = cmd.Start(); err != nil {
		fmt.Printf("Error: %s\n", err)
		log.Error("downFile error:%s, url:%s", err.Error(), strCmd)
	}

	errPipeContent, err := ioutil.ReadAll(errPipe)
	if err != nil {
		log.Error("read piple:%s,cmd:%s", err.Error(), strCmd)
	}

	stdOutContent, err := ioutil.ReadAll(stdoutPip)
	if err != nil {
		log.Error("read piple:%s,cmd:%s", err.Error(), strCmd)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("err:", err, string(errPipeContent))
		log.Error("downFile error:%s:%s,url:%s", err.Error(), string(errPipeContent), strCmd)
		return ""
	}

	fmt.Println("1:", string(stdOutContent))

	log.Debug("exec success:%s", strCmd)
	fmt.Println("exec success:", strCmd)

	return strings.ToLower(string(stdOutContent))
}

//执行系统命令字符串
//字符串中，不可以有连续2个以上的空格
func ExeCmd(strCmd string) bool {
	args := strings.Split(strCmd, " ")
	cmd := exec.Command(args[0], args[1:]...)

	//fmt.Println(strCmd)

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		log.Error("exeCmd:%v", err)
		return false
	}

	if err = cmd.Start(); err != nil {
		fmt.Printf("Error: %s\n", err)
		log.Error("downFile error:%s, url:%s", err.Error(), strCmd)
	}

	errPipeContent, err := ioutil.ReadAll(errPipe)
	if err != nil {
		log.Error("read piple:%s,cmd:%s", err.Error(), strCmd)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("err:", err, string(errPipeContent))
		log.Error("downFile error:%s:%s,url:%s", err.Error(), string(errPipeContent), strCmd)
		return false
	}

	log.Debug("exec wget success:\n%s", strCmd)

	return true
}

func DownLoadFile(inputUrl string) string {
	pos := strings.Index(inputUrl, " ")
	if pos != -1 {
		log.Debug("url:%s", inputUrl)
		inputUrl = utils.ClearSpecialChar(inputUrl)
	}

	path := utils.GenerateSmallFileName(inputUrl)
	if path == "" {
		log.Error("err path:%s", path)
		return ""
	}

	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		log.Error("%s", err.Error())
		fmt.Println(err)
		return ""
	}
	log.Debug("DownLoadFile src url:%s", inputUrl)
	log.Debug("DownLoadFile dst path:%s", path)
	_, err = os.Stat(path)
	if err == nil {
		os.Remove(path)
		log.Info("this file:%s,have exsisted.just delete it.", path)
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
		log.Error("err url:%s", inputUrl)
		return ""
	}

	log.Debug("down url:%s", strCmd)

	b := ExeCmd(strCmd)
	if !b {
		return ""
	}

	log.Debug("DownLoadFile success:%s", inputUrl)

	return path
}
