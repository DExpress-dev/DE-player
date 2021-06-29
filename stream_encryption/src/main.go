package main

/*
const char* build_time(void)
{
    static const char* psz_build_time = "["__DATE__ "  " __TIME__ "]";
    return psz_build_time;

}
*/
import "C"
import (
	"business"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	log4plus "log4go"
)

//版本号
var (
	ver       string = "1.0.2"
	exeName   string = "Download Encryption"
	buildTime string = C.GoString(C.build_time())
)

type Config struct {
	ChannelListen string `json:"ChannelListen"` //管理查询
	AdminListen   string `json:"AdminListen"`   //体感上报
}

type ConfigBusiness struct {
	adminListen   string
	channelListen string
	web           *business.WebManager //web管理类

}

type Flags struct {
	Help      bool
	Version   bool
	Key       string //input stream key
	StreamUrl string //input stream pull url
	LocalPath string //input stream download local path
	PushUrl   string //input stream push url
}

var gConfig ConfigBusiness

func (f *Flags) Init() {
	flag.BoolVar(&f.Help, "h", false, "help")
	flag.BoolVar(&f.Version, "V", false, "show version")
	flag.StringVar(&f.Key, "k", "", "first stream key")
	flag.StringVar(&f.StreamUrl, "u", "", "first stream url")
	flag.StringVar(&f.LocalPath, "l", "", "first stream local path ended with /")
	flag.StringVar(&f.PushUrl, "p", "", "first stream push url")
}

func (f *Flags) Check() (needReturn bool) {
	flag.Parse()

	if f.Help {
		flag.Usage()
		needReturn = true
	} else if f.Version {
		verString := exeName + " Version: " + ver + "\r\n"
		verString += "compile time:" + buildTime + "\r\n"
		fmt.Println(verString)
		needReturn = true
	}

	return needReturn
}

var flags *Flags = &Flags{}

func init() {
	flags.Init()
	exeName = getExeName()
}

func getExeName() string {
	ret := ""
	ex, err := os.Executable()
	if err == nil {
		ret = filepath.Base(ex)
	}
	return ret
}

func setLog() {
	logJson := "log.json"
	set := false
	if bExist := business.PathExist(logJson); bExist {
		if err := log4plus.SetupLogWithConf(logJson); err == nil {
			set = true
		}
	}

	if !set {
		fileWriter := log4plus.NewFileWriter()
		exeName := getExeName()
		fileWriter.SetPathPattern("./log/" + exeName + "-%Y%M%D.log")
		log4plus.Register(fileWriter)
		log4plus.SetLevel(log4plus.DEBUG)
	}
}

func (c *ConfigBusiness) configLoad() bool {

	//加载config.json 配置
	config := &Config{}
	cfgFile, err := os.Open("config.json")
	if err != nil {
		log4plus.Error("(c *ConfigBusiness) configLoad Failed Open File Error %s", err.Error())
		return false
	}
	defer cfgFile.Close()
	log4plus.Info("(c *ConfigBusiness) configLoad Open config.json Success")

	cfgBytes, _ := ioutil.ReadAll(cfgFile)
	jsonErr := json.Unmarshal(cfgBytes, config)
	if jsonErr != nil {
		log4plus.Error("(c *ConfigBusiness) configLoad json.Unmarshal Failed %s", jsonErr.Error())
		return false
	}
	gConfig.adminListen = config.AdminListen
	gConfig.channelListen = config.ChannelListen

	//显示加载的信息
	log4plus.Info("(c *ConfigBusiness) configLoad config.json \n adminListen=%s\n channelListen=%s\n", gConfig.adminListen, gConfig.channelListen)
	return true
}

func main() {

	needReturn := flags.Check()
	if needReturn {
		return
	}

	setLog()
	defer log4plus.Close()
	log4plus.Info("%s Version=%s Build Time=%s", getExeName(), ver, buildTime)
	defer log4plus.Close()

	//加载配置信息
	log4plus.Info("Start configLoad")
	if !gConfig.configLoad() {
		log4plus.Error("configLoad Failed Exit Program")
		return
	}

	//启动Web
	log4plus.Info("Create Web Manager")
	gConfig.web = business.New(gConfig.channelListen, gConfig.adminListen, "127.0.0.1")

	for {
		time.Sleep(time.Duration(1) * time.Second)
	}
}
