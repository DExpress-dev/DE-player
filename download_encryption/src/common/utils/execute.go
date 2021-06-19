package utils

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

//执行系统命令字符串
//字符串中，不可以有连续2个以上的空格
func ExeCmd(strCmd string) bool {
	args := strings.Split(strCmd, " ")
	cmd := exec.Command(args[0], args[1:]...)

	pipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return false
	}
	if err = cmd.Start(); err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	pipeContent, err := ioutil.ReadAll(pipe)
	if err != nil {
		fmt.Printf("read piple:%s,cmd:%s\n", err.Error(), strCmd)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("err:", err, string(pipeContent))
		return false
	}

	return true
}

//执行系统命令字符串
//字符串中，不可以有连续2个以上的空格
func ExeCmdAndStdOut(strCmd string) string {
	args := strings.Split(strCmd, " ")
	cmd := exec.Command(args[0], args[1:]...)

	stdoutPip, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("StdoutPipe Error: %s\n", err)
		return ""
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return ""
	}

	if err = cmd.Start(); err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	errPipeContent, _ := ioutil.ReadAll(errPipe)

	stdOutContent, _ := ioutil.ReadAll(stdoutPip)

	if err := cmd.Wait(); err != nil {
		fmt.Println("err:", err, string(errPipeContent))
		return ""
	}

	return strings.ToLower(string(stdOutContent))
}
