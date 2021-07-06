package download

import (
	"common/hls"
	"config"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"public"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	log4plus "log4go"
)

type StreamConfig struct {
	ChannelName    string `json:"channelName"`    //频道名称
	SourceUrl      string `json:"sourceUrl"`      //源流地址
	PushUrl        string `json:"pushUrl"`        //推送给第三方地址
	SrcPath        string `json:"srcPath"`        //源流保存地址
	Key            string `json:"key"`            //加密密钥
	IV             string `json:"iv"`             //加密向量
	EncryptionPath string `json:"encryptionPath"` //加密流保存地址
}

type StreamsConfig struct {
	Stream []StreamConfig `json:"streams"`
}

//得到文件的创建时间
type FileCreateTimer struct {
	filePath    string
	createTimer int64
}
type FileCreateTimers []FileCreateTimer

type StreamDownload struct {
	ChannelName     string
	StreamUrl       string
	PushUrl         string
	SrcPath         string
	Key             string
	IV              string
	EncryptionPath  string
	AntileechRemote string
	M3u8FailCount   int
	IndexM3u8Pushed bool
	MaxFileCount    int
	DeleteFileCount int
	HlsStream       *hls.Stream
	M3u8            *hls.M3u8
	LastM3u8Time    time.Time
}

func (s *StreamDownload) Download() {

	hlsStream := hls.NewStream(
		s.ChannelName,
		s.StreamUrl,
		s.SrcPath,
		s.Key,
		s.IV,
		s.EncryptionPath).
		SetTimeout(config.GetInstance().Config.Download.Timeout).
		SetRetryCount(config.GetInstance().Config.Download.RetryCount).
		SetRetryWait(config.GetInstance().Config.Download.RetryWait).
		SetCallback(hls.StreamCallback{
			s.HandleError,
			s.HandleParseM3u8,
			s.HandleDownloadedM3u8,
			s.HandleDownloadedTs,
		})

	m3u8, err := hlsStream.DownloadM3u8()
	if err != nil {
		log4plus.Error("[%s]Download DownloadM3u8 err=%s ulr=%s timeout=%d", s.ChannelName, err.Error(), hlsStream.M3u8Url, hlsStream.Timeout)
		return
	}

	//判断是1级m3u8还是2级m3u8
	if m3u8.IsTop() {
		s.M3u8 = m3u8
		remotePath := hlsStream.M3u8UrlInfo.Scheme + "://" + hlsStream.M3u8UrlInfo.Host + path.Dir(hlsStream.M3u8UrlInfo.Path)
		bandMax, bandMin := m3u8.M3u8Entries[0], m3u8.M3u8Entries[0]

		for _, entry := range m3u8.M3u8Entries {
			if entry.Bandwidth > bandMax.Bandwidth {
				bandMax = entry
			}
			if entry.Bandwidth < bandMin.Bandwidth {
				bandMin = entry
			}
		}

		if "max" == config.GetInstance().Config.Stream.BandWidth {
			streamPath := bandMax.Raw
			if !strings.HasPrefix(streamPath, "http") {
				streamPath = remotePath + "/" + streamPath
			}
			hlsStream = hlsStream.AddStream(streamPath)
		} else if "min" == config.GetInstance().Config.Stream.BandWidth {
			streamPath := bandMin.Raw
			if !strings.HasPrefix(streamPath, "http") {
				streamPath = remotePath + "/" + streamPath
			}
			hlsStream = hlsStream.AddStream(streamPath)
		} else if "all" == config.GetInstance().Config.Stream.BandWidth {

		} else {
			band, _ := strconv.ParseInt(config.GetInstance().Config.Stream.BandWidth, 10, 64)
			for _, entry := range m3u8.M3u8Entries {
				if band == entry.Bandwidth {
					streamPath := entry.Raw
					if !strings.HasPrefix(streamPath, "http") {
						streamPath = remotePath + "/" + streamPath
					}
					hlsStream = hlsStream.AddStream(streamPath)
				}
			}
		}
	}
	s.HlsStream = hlsStream
	hlsStream.Pull()

	for {
		if s.M3u8 != nil && ((!s.HlsStream.IsTop) || (s.HlsStream.IsTop && !s.IndexM3u8Pushed)) {
			if s.HlsStream.IsTop {
				s.IndexM3u8Pushed = true
			}
			// go udp.SendM3u8(s.Key, s.LocalPath, s.M3u8)
		}

		//check second m3u8 update
		if s.HlsStream != nil && !s.HlsStream.IsTop {
			if time.Now().Sub(s.LastM3u8Time) > time.Minute {
				log4plus.Warn("[%s]Download m3u8 not changed over 1 minute!", s.ChannelName)
			}
		}

		time.Sleep(30 * time.Second)
	}
}

func (s *StreamDownload) getFileCreateTime(path string) int64 {

	osType := runtime.GOOS
	if fileInfo, err := os.Stat(path); err == nil {
		if osType == "linux" {
			stat_t := fileInfo.Sys().(*syscall.Stat_t)
			tCreate := int64(stat_t.Ctimespec.Sec) //linux 用 Ctim；
			/*windows 用
			wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
			tNanSeconds := wFileSys.CreationTime.Nanoseconds() /// 返回的是纳秒
			tSec := tNanSeconds / 1e9                          ///秒 */

			return tCreate
		}
	}
	return time.Now().Unix()
}

func (f FileCreateTimers) Len() int {
	return len(f)
}
func (f FileCreateTimers) Less(i, j int) bool {
	return f[i].createTimer < f[j].createTimer
}
func (f FileCreateTimers) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (s *StreamDownload) getAllFiles(folder string) []string {

	var filesArray []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		filesArray = append(filesArray, path)
		return nil
	})
	if err != nil {
		log4plus.Error("[%s] getAllFiles Failed Folder %s", s.ChannelName, folder)
		return filesArray
	}

	var files FileCreateTimers
	for _, v := range filesArray {
		singleFileCreateTimer := s.getFileCreateTime(v)
		file := FileCreateTimer{
			filePath:    v,
			createTimer: singleFileCreateTimer,
		}
		files = append(files, file)
	}

	//数组排序
	sort.Sort(files)

	//输出排序后的文件数组
	var fileAllArray []string
	for _, v := range files {
		fileAllArray = append(fileAllArray, v.filePath)
	}
	return fileAllArray
}

func (s *StreamDownload) Delete() {

	for {
		time.Sleep(5 * time.Second)

		//删除源流
		srcFiles := s.getAllFiles(s.SrcPath)
		if len(srcFiles) > s.MaxFileCount {

			for i := 0; i <= s.DeleteFileCount; i++ {
				deleteFile := srcFiles[0]
				srcFiles = append(srcFiles[:0], srcFiles[1:]...)
				os.Remove(deleteFile)
			}

		}

		//删除加密流
		encryptionFiles := s.getAllFiles(s.EncryptionPath)
		if len(encryptionFiles) > s.MaxFileCount {

			for i := 0; i <= s.DeleteFileCount; i++ {
				deleteFile := encryptionFiles[0]
				encryptionFiles = append(encryptionFiles[:0], encryptionFiles[1:]...)
				os.Remove(deleteFile)
			}
		}
	}
}

func (s *StreamDownload) HandleError(hlsStream *hls.Stream, err hls.Error) {

	//	switch err.Code {
	//	case hls.ErrorCodeM3u8DownloadFail, hls.ErrorCodeM3u8FormatError:
	//		s.M3u8FailCount++
	//		if s.M3u8FailCount >= 3 {
	//			log4plus.Error("[%s]OnStreamError err=%s m3u8Url=%s failCount=%d", hlsStream.ChannelName, err.Err, hlsStream.M3u8Url, s.M3u8FailCount)
	//		} else {
	//			log4plus.Warn("[%s]OnStreamError err=%s m3u8Url=%s failCount=%d", hlsStream.ChannelName, err.Err, hlsStream.M3u8Url, s.M3u8FailCount)
	//		}
	//	case hls.ErrorCodeTsDownloadFail:
	//		ts := err.Data.(*hls.Ts)
	//		if ts.Status == hls.TsStatusFail {
	//			log4plus.Error("[%s]OnStreamError err=%s tsUrl=%s", hlsStream.ChannelName, err.Err, ts.TsUrl)
	//		}
	//	default:
	//		log4plus.Error("[%s]OnStreamError err=%s m3u8Url=%s", hlsStream.ChannelName, err.Err, hlsStream.M3u8Url)
	//	}
}

func (s *StreamDownload) HandleParseM3u8(hlsStream *hls.Stream, m3u8 *hls.M3u8) {
	log4plus.Debug("[%s]HandleParseM3u8 localFile=%s", hlsStream.ChannelName, m3u8.LocalFile)
	s.M3u8FailCount = 0
	s.LastM3u8Time = time.Now()
}

func (s *StreamDownload) HandleDownloadedM3u8(hlsStream *hls.Stream, m3u8 *hls.M3u8) {
	log4plus.Debug("[%s]HandleDownloadedM3u8 uri=%s localFile=%s", hlsStream.ChannelName, m3u8.M3u8Url, m3u8.LocalFile)
}

func (s *StreamDownload) HandleDownloadedTs(hlsStream *hls.Stream, ts *hls.Ts) {
	log4plus.Debug("[%s]HandleDownloadedTs uri=%s localFile=%s", hlsStream.ChannelName, ts.TsUrl, ts.SrcFile)
}

type StreamsDownload struct {
	Streams         *StreamsConfig
	MaxFileCount    int
	DeleteFileCount int
	streamMap       map[string]*StreamDownload
}

func (server *StreamsDownload) StreamAdd(channelName, streamUrl, pushUrl, srcPath, key, iv, encryptionPath string) *StreamDownload {

	if channelName == "" || streamUrl == "" || (pushUrl == "" && srcPath == "") {
		return nil
	}

	if srcPath != "" {
		os.MkdirAll(srcPath, os.ModePerm)
	}

	if key != "" && encryptionPath != "" {
		os.MkdirAll(encryptionPath, os.ModePerm)
	}

	stream := &StreamDownload{
		ChannelName:     channelName,
		StreamUrl:       streamUrl,
		PushUrl:         pushUrl,
		SrcPath:         srcPath,
		Key:             key,
		IV:              iv,
		EncryptionPath:  encryptionPath,
		LastM3u8Time:    time.Now(),
		IndexM3u8Pushed: false,
		MaxFileCount:    server.MaxFileCount,
		DeleteFileCount: server.DeleteFileCount,
	}
	server.streamMap[channelName] = stream
	return stream
}

func (server *StreamsDownload) loadConfig() {

	log4plus.Info("loadConfig")

	currentPath := public.GetCurrentDirectory()
	urlPath := currentPath + "/" + config.GetInstance().Config.Stream.UrlFile
	if !public.FileExist(urlPath) {
		return
	}

	server.Streams = &StreamsConfig{}
	cfgFile, err := os.Open(urlPath)
	if err != nil {
		log4plus.Error("loadConfig Failed Open File Error %s", err.Error())
		return
	}
	defer cfgFile.Close()
	log4plus.Info("loadConfig Open config.json Success")

	cfgBytes, _ := ioutil.ReadAll(cfgFile)
	jsonErr := json.Unmarshal(cfgBytes, server.Streams)
	if jsonErr != nil {
		log4plus.Error("loadConfig json.Unmarshal Failed %s", jsonErr.Error())
		return
	}

	for _, v := range server.Streams.Stream {
		server.StreamAdd(v.ChannelName, v.SourceUrl, v.PushUrl, v.SrcPath, v.Key, v.IV, v.EncryptionPath)
	}
}

func New() *StreamsDownload {

	server := &StreamsDownload{
		MaxFileCount:    config.GetInstance().Config.Stream.MaxFileCount,
		DeleteFileCount: config.GetInstance().Config.Stream.DeleteCount,
		streamMap:       make(map[string]*StreamDownload),
	}
	server.loadConfig()
	return server
}

func (server *StreamsDownload) Run() {

	//流下载
	log4plus.Info("StreamsDownload Current Stream Count=%d...", len(server.streamMap))
	for _, stream := range server.streamMap {
		go stream.Download()
	}

	//文件删除
	for _, stream := range server.streamMap {
		go stream.Delete()
	}
}
