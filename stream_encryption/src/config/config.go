package config

import (
	log4plus "common/log4go"
	"encoding/json"
	"io/ioutil"
	"os"

	_ "github.com/widuu/goini"
)

type StreamConfig struct {
	Url          string `json:"url"`
	UrlFile      string `json:"url_file"`
	BandWidth    string `json:"bandwidth"`
	MaxFileCount int    `json:"max_count"`
	DeleteCount  int    `json:"delete_count"`
}

type DownloadConfig struct {
	Timeout    int `json:"timeout"`
	RetryCount int `json:"retry_count"`
	RetryWait  int `json:"retry_wait"`
}

type UDPConfig struct {
	Remote string `json:"remote"`
	Folder string `json:"folder"`
}

type ConfigParam struct {
	Admin    string         `json:"admin"`
	Stream   StreamConfig   `json:"stream"`
	Download DownloadConfig `json:"download"`
	Udp      UDPConfig      `json:"udp"`
}

type configInfo struct {
	Config *ConfigParam
}

var _cfg *configInfo

func GetInstance() *configInfo {
	return _cfg
}

func init() {
	_cfg = new(configInfo)

	//加载config.json 配置
	_cfg.Config = &ConfigParam{}
	cfgFile, err := os.Open("config.json")
	if err != nil {
		log4plus.Error("configLoad Failed Open File Error %s", err.Error())
		return
	}
	defer cfgFile.Close()
	log4plus.Info("configLoad Open config.json Success")

	cfgBytes, _ := ioutil.ReadAll(cfgFile)
	jsonErr := json.Unmarshal(cfgBytes, _cfg.Config)
	if jsonErr != nil {
		log4plus.Error("configLoad json.Unmarshal Failed %s", jsonErr.Error())
		return
	}
}
