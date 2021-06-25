package udp

import (
	"bytes"
	"common/hls"
	"common/utils"
	"config"
	"fmt"
	"net/http"
	"path"
	"public"
	"sync"
	"time"

	log4plus "common/log4go"
)

const (
	SendTsArrayMax = 200
	SendTsArrayMin = 100
)

type UdpManager struct {
	lock        sync.Mutex
	sendTsMap   map[string]bool
	sendTsArray []string
}

var udpManager *UdpManager

func init() {
	udpManager = NewUdpManager()
}

func NewUdpManager() *UdpManager {
	return &UdpManager{sendTsMap: make(map[string]bool)}
}

func (udpManager *UdpManager) tsArrayShrink() {
	tsLen := len(udpManager.sendTsArray)
	if tsLen < SendTsArrayMax {
		return
	}

	for i := 0; i < tsLen-SendTsArrayMin; i++ {
		delete(udpManager.sendTsMap, udpManager.sendTsArray[i])
	}

	udpManager.sendTsArray = udpManager.sendTsArray[tsLen-SendTsArrayMin:]
}

func (udpManager *UdpManager) NeedSend(ts *hls.Ts) bool {
	udpManager.lock.Lock()
	defer udpManager.lock.Unlock()

	udpManager.tsArrayShrink()
	if udpManager.sendTsMap[ts.Name] == true {
		return false
	}
	udpManager.sendTsMap[ts.Name] = true
	udpManager.sendTsArray = append(udpManager.sendTsArray, ts.Name)
	return true
}

func SendHls(key, path, rename, extra string) error {
	fileSize, err := utils.FileSize(path)
	if err != nil {
		return fmt.Errorf("%s path=%s", err.Error(), path)
	}
	if fileSize == 0 {
		return fmt.Errorf("file size = 0 path=%s", path)
	}

	bodyFmt := `
{
    "protocol": 0,
    "hls": {
        "path": "%s",
        "rename": "%s",
        "extra": "%s"
    }
}`
	body := bytes.NewBufferString(fmt.Sprintf(bodyFmt, path, rename, extra))
	log4plus.Debug("[%s]SendHls path=%s rename=%s extra=%s fileSize=%d Remote=%s", key, path, rename, extra, fileSize, config.GetInstance().Config.Udp.Remote)
	_, err = http.Post(config.GetInstance().Config.Udp.Remote, "application/json", body)
	if err != nil {
		// log4plus.Error("[%s]SendHls http.Post err=%s remote=%s", key, err.Error(), config.GetInstance().Config.Udp.Remote)
		return err
	}

	return nil
}

func SaveSecondM3u8(localFile, destPath string) (destFile string, err error) {
	fileName := path.Base(localFile)
	prePath := time.Now().Format("2006010215")
	destFile = destPath + "/" + prePath + "/" + fileName + "." + time.Now().Format("20060102150405")
	err = public.CopyFile(localFile, destFile)
	if err != nil {
		return "", err
	}
	return destFile, nil
}

func SendTs(key, localPath string, ts *hls.Ts) error {
	if !udpManager.NeedSend(ts) {
		return nil
	}

	extraPrefix := path.Base(localPath)
	err := SendHls(key, ts.LocalFile, ts.Name, extraPrefix+"/"+ts.RelativeToParent)
	if err != nil {
		// log4plus.Error("[%s]SendTs SendHls err=%s", key, err.Error())
		return err
	}
	return nil
}

func SendM3u8(key, localPath string, m3u8 *hls.M3u8) (err error) {
	extraPrefix := path.Base(localPath)
	fileName := path.Base(m3u8.LocalFile)

	if m3u8.IsTop() {
		err = SendHls(key, m3u8.LocalFile, fileName, extraPrefix)
		if err != nil {
			log4plus.Error("[%s]udp.SendM3u8 SendHls err=%s", key, err.Error())
		}
	} else {
		udpPath := localPath + "/" + config.GetInstance().Config.Udp.Folder
		destFile, err := SaveSecondM3u8(m3u8.LocalFile, udpPath)
		if err != nil {
			log4plus.Error("[%s]udp.SendM3u8 SaveSecondM3u8 err=%s localFile=%s", key, err.Error(), m3u8.LocalFile)
			return err
		}
		err = SendHls(key, destFile, fileName, extraPrefix)
		if err != nil {
			// log4plus.Error("[%s]udp.SendM3u8 SendHls err=%s", key, err.Error())
		}
	}

	return nil
}

func SendRtmp(source, dest string) error {
	bodyFmt := `
{
    "protocol": 1,
    "rtmp": {
        "source": "%s",
        "dest": "%s"
    }
}`
	body := bytes.NewBufferString(fmt.Sprintf(bodyFmt, source, dest))
	log4plus.Debug("SendRtmp source=%s dest=%s", source, dest)

	_, err := http.Post(config.GetInstance().Config.Udp.Remote, "application/json", body)
	if err != nil {
		log4plus.Error("SendRtmp http.Post err=%s remote=%s", err.Error(), config.GetInstance().Config.Udp.Remote)
		return err
	}

	return nil
}
