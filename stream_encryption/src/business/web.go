package business

import (
	"config"
	"download"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"public"
	_ "strconv"
	_ "strings"
	"sync"
	"time"

	log4plus "log4go"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

const (
	StatusJson = 600 // 解析Json格式失败
	StatusIp   = 601 // 解析Json格式失败
)

const (
	StatusJsonString = "" // 解析Json格式失败
)

type ResponseResult struct {
	Result  int    `json:"result"`
	Message string `json:"message"`
}

//添加流
type AddStreamRequest struct {
	ChannelName    string `json:"channelName"`    //频道名称
	SourceUrl      string `json:"sourceUrl"`      //频道地址
	PushUrl        string `json:"pushUrl"`        //推送地址
	SrcPath        string `json:"SrcPath"`        //本地保存地址
	Key            string `json:"key"`            //加密密钥
	IV             string `json:"iv"`             //加密向量
	EncryptionPath string `json:"encryptionPath"` //加密密钥
}

//删除流
type DeleteStreamRequest struct {
	ChannelName string `json:"channelName"` //频道名称
	DeleteLocal bool   `json:"deleteLocal"` //频道名称
}

type AdminsIP struct {
	Admins []string `json:"admins"`
}

type StreamConfigPath struct {
	SrcPath        string
	EncryptionPath string
}

type WebManager struct {
	ChannelListen string
	AdminListen   string
	Admins        map[string]bool

	streamLock     sync.Mutex
	streamsConfig  map[string]*StreamConfigPath
	streamDownload *download.StreamsDownload

	ChannelGin *gin.Engine
	AdminGin   *gin.Engine
}

var (
	logFilePath = "./"
	logFileName = "download_encryption.log"
)

//解决跨域问题
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "PUT, DELETE, POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}

func logerMiddleware() gin.HandlerFunc {
	// 日志文件
	fileName := path.Join(logFilePath, logFileName)
	// 写入文件
	//src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	src, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("err", err)
	}
	// 实例化
	logger := logrus.New()
	//设置日志级别
	logger.SetLevel(logrus.DebugLevel)
	//设置输出
	logger.Out = src
	// 设置 rotatelogs
	logWriter, err := rotatelogs.New(
		// 分割后的文件名称
		fileName+".%Y%m%d.log",

		// 生成软链，指向最新日志文件
		rotatelogs.WithLinkName(fileName),

		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(7*24*time.Hour),

		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}

	logger.AddHook(lfshook.NewHook(writeMap, &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
	}))

	return func(c *gin.Context) {
		//开始时间
		startTime := time.Now()
		//处理请求
		c.Next()
		//结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		//请求方式
		reqMethod := c.Request.Method
		//请求路由
		reqUrl := c.Request.RequestURI
		//状态码
		statusCode := c.Writer.Status()
		//请求ip
		clientIP := c.ClientIP()

		// 日志格式
		logger.WithFields(logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUrl,
		}).Info()

	}
}

func (wm *WebManager) GetClientIP(c *gin.Context) string {
	reqIP := c.ClientIP()
	if reqIP == "::1" {
		reqIP = "127.0.0.1"
	}
	return reqIP
}

func (wm *WebManager) findAdmin(Ip string) bool {

	exist := wm.Admins[Ip]
	return exist
}

func (wm *WebManager) checkIp(c *gin.Context) bool {

	clientIp := wm.GetClientIP(c)
	if wm.findAdmin(clientIp) {

		log4plus.Info("---->>>>checkIp %s****", clientIp)
		return true
	}
	return false
}

func (wm *WebManager) GetIndexStream(c *gin.Context) {

	//获取到客户端IP地址
	clientIP := wm.GetClientIP(c)
	log4plus.Info("---->>>>GetIndexStream Request clientIP=%s****", clientIP)

	//获取到频道名
	channelName := c.Param("channelName")
	dirName := c.Param("dirName")

	//判断流是否存在
	srcPath, encryptionPath, err := wm.findStream(channelName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"result":  http.StatusNotFound,
			"message": "Stream Not Found",
		})
		return
	}

	//检测IP地址
	var srcFile string
	if wm.checkIp(c) {
		srcFile = srcPath + "/" + dirName
	} else {
		srcFile = encryptionPath + "/" + dirName
	}

	if !public.FileExist(srcFile) {

		c.JSON(http.StatusNotFound, gin.H{
			"result":  http.StatusNotFound,
			"message": "FileName Not Found",
		})
		return
	}

	//下发文件
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+dirName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(srcFile)
}

func (wm *WebManager) GetTsStream(c *gin.Context) {

	//获取到客户端IP地址
	clientIP := wm.GetClientIP(c)
	log4plus.Info("---->>>>GetTsStream Request clientIP=%s****", clientIP)

	//获取到频道名
	channelName := c.Param("channelName")
	dirName := c.Param("dirName")
	fileName := c.Param("fileName")

	//判断流是否存在
	srcPath, encryptionPath, err := wm.findStream(channelName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"result":  http.StatusNotFound,
			"message": "Stream Not Found",
		})
		return
	}

	//检测IP地址
	var srcFile string
	if wm.checkIp(c) {
		srcFile = srcPath + "/" + dirName + "/" + fileName
	} else {
		srcFile = encryptionPath + "/" + dirName + "/" + fileName
	}

	if !public.FileExist(srcFile) {

		c.JSON(http.StatusNotFound, gin.H{
			"result":  http.StatusNotFound,
			"message": "FileName Not Found",
		})
		return
	}

	//下发文件
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(srcFile)
}

//*********************************
//此处约定，凡是key不为空的，都需要进行流加密
func (wm *WebManager) AddStream(c *gin.Context) {

	log4plus.Info("---->>>>AddStream****")

	//检测IP地址
	if !wm.checkIp(c) {
		c.JSON(StatusIp, gin.H{
			"result":  StatusIp,
			"message": "Client Ip not Admin",
		})
	}

	//分解数据
	var request AddStreamRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(StatusIp, gin.H{
			"result":  StatusIp,
			"message": "Unmarshal Add Stream Boby Json Failed",
		})
		return
	}

	//添加流
	wm.streamDownload.StreamAdd(request.ChannelName,
		request.SourceUrl,
		request.PushUrl,
		request.SrcPath,
		request.Key,
		request.IV,
		request.EncryptionPath)

	//
	c.JSON(http.StatusOK, gin.H{
		"result":  http.StatusOK,
		"message": "OK",
	})
}

func (wm *WebManager) DeleteStream(c *gin.Context) {

	log4plus.Info("---->>>>DeleteStream****")

	//检测IP地址
	if !wm.checkIp(c) {
		c.JSON(StatusIp, gin.H{
			"result":  StatusIp,
			"message": "Client Ip not Admin",
		})
	}

	//分解数据
	var request AddStreamRequest
	if err := c.BindJSON(&request); err != nil {

		c.JSON(StatusIp, gin.H{
			"result":  StatusIp,
			"message": "Unmarshal Add Stream Boby Json Failed",
		})
		return
	}

	//添加流
	wm.streamDownload.StreamAdd(request.ChannelName, request.SourceUrl, request.PushUrl, request.SrcPath, request.Key, request.IV, request.EncryptionPath)

	//
	c.JSON(http.StatusOK, gin.H{
		"result":  http.StatusOK,
		"message": "OK",
	})
}

func (wm *WebManager) ClearStream(c *gin.Context) {

	log4plus.Info("---->>>>ClearStream****")

	//检测IP地址
	if !wm.checkIp(c) {
		c.JSON(StatusIp, gin.H{
			"result":  StatusIp,
			"message": "Client Ip not Admin",
		})
	}
}

func (wm *WebManager) findStream(channelName string) (string, string, error) {

	wm.streamLock.Lock()
	defer wm.streamLock.Unlock()

	return wm.streamsConfig[channelName].SrcPath, wm.streamsConfig[channelName].EncryptionPath, nil
}

func (wm *WebManager) getStreamConfig() {

	log4plus.Info("getStreamConfig")

	currentPath := public.GetCurrentDirectory()
	urlPath := currentPath + "/" + config.GetInstance().Config.Stream.UrlFile

	if !public.FileExist(urlPath) {
		return
	}

	var streamsConfig download.StreamsConfig
	cfgFile, err := os.Open(urlPath)
	if err != nil {
		log4plus.Error("loadConfig Failed Open File Error %s", err.Error())
		return
	}
	defer cfgFile.Close()
	log4plus.Info("loadConfig Open config.json Success")

	cfgBytes, _ := ioutil.ReadAll(cfgFile)
	jsonErr := json.Unmarshal(cfgBytes, &streamsConfig)
	if jsonErr != nil {
		log4plus.Error("loadConfig json.Unmarshal Failed %s", jsonErr.Error())
		return
	}

	for _, v := range streamsConfig.Stream {
		var channelInfo *StreamConfigPath = new(StreamConfigPath)
		channelInfo.SrcPath = v.SrcPath
		channelInfo.EncryptionPath = v.EncryptionPath
		wm.streamsConfig[v.ChannelName] = channelInfo
	}
}

func (wm *WebManager) getAdminsConfig() {

	log4plus.Info("getAdminsConfig")

	currentPath := public.GetCurrentDirectory()
	urlPath := currentPath + "/" + config.GetInstance().Config.Admin

	if !public.FileExist(urlPath) {
		return
	}

	var adminsConfig AdminsIP
	cfgFile, err := os.Open(urlPath)
	if err != nil {
		log4plus.Error("getAdminsConfig Failed Open File Error %s", err.Error())
		return
	}
	defer cfgFile.Close()
	log4plus.Info("getAdminsConfig Open config.json Success")

	cfgBytes, _ := ioutil.ReadAll(cfgFile)
	jsonErr := json.Unmarshal(cfgBytes, &adminsConfig)
	if jsonErr != nil {
		log4plus.Error("getAdminsConfig json.Unmarshal Failed %s", jsonErr.Error())
		return
	}

	for _, v := range adminsConfig.Admins {
		wm.Admins[v] = true
	}
}

//流接口
func (wm *WebManager) startChannel() {

	//分组处理
	channelsGroup := wm.ChannelGin.Group("/channellist")
	{
		channelsGroup.GET("/:channelName/:dirName", wm.GetIndexStream)
		channelsGroup.GET("/:channelName/:dirName/:fileName", wm.GetTsStream)
	}
	wm.ChannelGin.Run(wm.ChannelListen)
}

//管理接口
func (wm *WebManager) startAdmin() {

	//分组处理
	adminGroup := wm.ChannelGin.Group("/admin")
	{
		adminGroup.POST("/addStream", wm.AddStream)
		adminGroup.POST("/deleteStream", wm.DeleteStream)
		adminGroup.POST("/clearStream", wm.ClearStream)
	}
	wm.AdminGin.Run(wm.AdminListen)
}

func New(channelListen string, adminListen string, adminIp string) *WebManager {

	//创建对象
	web := &WebManager{
		ChannelListen: channelListen,
		AdminListen:   adminListen,
		Admins:        make(map[string]bool),
		streamsConfig: make(map[string]*StreamConfigPath),
	}

	//获取流配置信息
	web.getStreamConfig()
	//获取admin的信息
	web.getAdminsConfig()

	//启动管理频道gin
	log4plus.Info("Create Channel Web Manager")
	web.ChannelGin = gin.Default()
	gin.SetMode(gin.ReleaseMode)
	web.ChannelGin.Use(logerMiddleware())
	web.ChannelGin.Use(Cors())

	//启动管理web gin
	log4plus.Info("Create Admin Web Manager")
	web.AdminGin = gin.Default()
	gin.SetMode(gin.ReleaseMode)
	web.AdminGin.Use(logerMiddleware())
	web.AdminGin.Use(Cors())

	//启动流下载
	log4plus.Info("Create Streams Download Manager")
	web.streamDownload = download.New()
	web.streamDownload.Run()

	//启动Web
	go web.startChannel()
	go web.startAdmin()
	return web
}
