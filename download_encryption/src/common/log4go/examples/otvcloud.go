// test
package main

import (
	"common/utils"
	"fmt"
	log "log4go"
	"os"
	"path/filepath"
)

//get app name
func getExeName() string {
	ret := ""
	ex, err := os.Executable()
	if err == nil {
		ret = filepath.Base(ex)
	}
	return ret
}

//open log engine
func setLogDefault() {
	//new  logWriter
	fileWriter := log.NewFileWriter()
	consoleWriter := log.NewConsoleWriter()

	//set  pattern
	exeName := getExeName()
	fileWriter.SetPathPattern("/var/log/go/" + exeName + "/" + exeName + "-%Y%M%D.log")
	//register
	log.Register(fileWriter)
	log.Register(consoleWriter)

	//set log level
	log.SetLevel(log.DEBUG)
}

func main() {
	//if no log.json, set log default ; if have log.json,  read log.json
	logJson := "log.json"
	bExist, _ := utils.PathExist(logJson)
	if !bExist {
		fmt.Println("no log.json")
		setLogDefault()
	} else {
		fmt.Println("read log.json")
		if err := log.SetupLogWithConf(logJson); err != nil {
			setLogDefault()
		}
	}

	defer log.Close()

	//write log example
	var name = "skoo"
	log.Debug("log4go by %s", name)
	log.Info("log4go by %s", name)
	log.Warn("log4go by %s", name)
	log.Error("log4go by %s", name)
	log.Fatal("log4go by %s", name)
}
