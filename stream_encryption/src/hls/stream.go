package hls

import (
	"antileech"
	"bytes"
	"common/httputils"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	StreamTsCountMax    = 200 //ts存放的数量;
	StreamTsCountReduce = 100 //一次性删除的数量;
)

const (
	ErrorCodeBase = iota + 1000
	ErrorCodeM3u8DownloadFail
	ErrorCodeM3u8FormatError
	ErrorCodeTsDownloadRetry
	ErrorCodeTsDownloadFail
	ErrorCodeStreamDisconnected
	ErrorCodeStreamRecover
	ErrorCodeM3u8SequenceNotContinuous
	ErrorCodeM3u8TsRepeat
	ErrorCodeTsDurationAbnormal
	ErrorCodeTsSizeZero
	ErrorCodeTsDurationMismatch
)

type Error struct {
	Code int
	Data interface{}
	Err  string
}

type OnError func(stream *Stream, err Error)
type OnParseM3u8 func(stream *Stream, m3u8 *M3u8)
type OnDownloadedM3u8 func(stream *Stream, m3u8 *M3u8)
type OnDownloadedTs func(stream *Stream, ts *Ts)

type StreamCallback struct {
	HandleError          OnError
	HandleParseM3u8      OnParseM3u8
	HandleDownloadedM3u8 OnDownloadedM3u8
	HandleDownloadedTs   OnDownloadedTs
}

type Stream struct {
	Key             string
	IV              string
	M3u8Url         string
	M3u8UrlInfo     *url.URL
	M3u8Name        string
	M3u8String      string
	LocalPath       string
	LocalFile       string
	Timeout         int
	RetryCount      int
	RetryWait       int
	LastSequence    int64
	AntileechRemote string
	OnlyM3u8        bool
	SaveM3u8        bool
	SaveTs          bool
	IsTop           bool
	Lock            sync.Mutex
	Stoped          bool
	Closed          chan struct{}

	//top m3u8
	Streams map[string]*Stream

	//second m3u8
	TsMap map[string]*Ts

	//callbacks
	Callback StreamCallback
}

func NewStream(key, m3u8Url, localPath string) *Stream {

	localPath = strings.TrimSpace(localPath)
	if strings.HasSuffix(localPath, "/") {
		localPath = localPath[:len(localPath)-1]
	}
	stream := &Stream{
		Key:          key,
		IV:           key,
		M3u8Url:      m3u8Url,
		LocalPath:    localPath,
		LastSequence: -1,
		Timeout:      5,
		RetryCount:   3,
		RetryWait:    2,
		OnlyM3u8:     false,
		SaveM3u8:     true,
		SaveTs:       true,
		Streams:      make(map[string]*Stream),
		TsMap:        make(map[string]*Ts),
	}

	stream.M3u8UrlInfo, _ = url.Parse(m3u8Url)
	stream.M3u8Name = path.Base(stream.M3u8UrlInfo.Path)
	if localPath != "" {
		stream.LocalFile = localPath + "/" + stream.M3u8Name
	}

	return stream
}

func (stream *Stream) SetCallback(callback StreamCallback) *Stream {
	stream.Callback = callback
	return stream
}

func (stream *Stream) SetTimeout(timeout int) *Stream {
	stream.Timeout = timeout
	return stream
}

func (stream *Stream) SetRetryCount(retryCount int) *Stream {
	stream.RetryCount = retryCount
	return stream
}

func (stream *Stream) SetRetryWait(retryWait int) *Stream {
	stream.RetryWait = retryWait
	return stream
}

func (stream *Stream) SetAntileechRemote(antileechRemote string) *Stream {
	stream.AntileechRemote = antileechRemote
	return stream
}

func (stream *Stream) SetSaveM3u8(saveM3u8 bool) *Stream {
	stream.SaveM3u8 = saveM3u8
	return stream
}

func (stream *Stream) SetSaveTs(saveTs bool) *Stream {
	stream.SaveTs = saveTs
	return stream
}

func (stream *Stream) SetOnlyM3u8(onlyM3u8 bool) *Stream {
	stream.OnlyM3u8 = onlyM3u8
	return stream
}

func (stream *Stream) handleError(code int, data interface{}, format string, args ...interface{}) {
	if stream.Callback.HandleError != nil {
		go stream.Callback.HandleError(stream, Error{code, data, fmt.Sprintf(format, args...)})
	}
}

func (stream *Stream) handleParseM3u8(m3u8 *M3u8) {
	if stream.Callback.HandleParseM3u8 != nil {
		go stream.Callback.HandleParseM3u8(stream, m3u8)
	}
}

func (stream *Stream) handleDownloadedTs(ts *Ts) {
	if stream.Callback.HandleDownloadedTs != nil {
		go stream.Callback.HandleDownloadedTs(stream, ts)
	}
}

func (stream *Stream) handleDownloadedM3u8(m3u8 *M3u8) {
	if stream.Callback.HandleDownloadedM3u8 != nil {
		go stream.Callback.HandleDownloadedM3u8(stream, m3u8)
	}
}

func (stream *Stream) FindStream(m3u8Name string) *Stream {
	stream.Lock.Lock()
	defer stream.Lock.Unlock()

	return stream.Streams[m3u8Name]
}

func (stream *Stream) AddStream(m3u8Url string) *Stream {
	stream.Lock.Lock()
	defer stream.Lock.Unlock()
	m3u8UrlInfo, _ := url.Parse(m3u8Url)
	stream.Streams[m3u8UrlInfo.Path] = NewStream(stream.Key, m3u8Url, stream.LocalPath).
		SetTimeout(stream.Timeout).
		SetRetryCount(stream.RetryCount).
		SetRetryWait(stream.RetryWait).
		SetSaveM3u8(stream.SaveM3u8).
		SetSaveTs(stream.SaveTs).
		SetOnlyM3u8(stream.OnlyM3u8).
		SetAntileechRemote(stream.AntileechRemote).
		SetCallback(stream.Callback)
	return stream.Streams[m3u8UrlInfo.Path]
}

func (stream *Stream) doDownloadTs(tsUrl, tsLocalFile string) (size int64, err error) {

	stream.handleError(ErrorCodeTsDownloadRetry, nil, "doDownloadTs--->>> tsUrl=%s tsLocalFile=%s", tsUrl, tsLocalFile)

	if stream.SaveTs && tsLocalFile != "" {
		size, err = httputils.DownloadFile(tsUrl, tsLocalFile, stream.Timeout)
	} else {
		buf := new(bytes.Buffer)
		size, err = httputils.DownloadBuffer(tsUrl, stream.Timeout, buf)
	}
	stream.handleError(ErrorCodeTsDownloadRetry, nil, "<<<---doDownloadTs tsUrl=%s tsLocalFile=%s", tsUrl, tsLocalFile)

	if err != nil {
		return 0, err
	}
	return size, nil
}

func (stream *Stream) downloadTs(ts *Ts, results chan<- *Ts) {

	if ts.UrlInfo.RawQuery != "" {
		ts.TsUrl = stream.M3u8UrlInfo.Scheme + "://" + stream.M3u8UrlInfo.Host + path.Dir(stream.M3u8UrlInfo.Path) + "/" + ts.RelativeToParent + "/" + ts.Name + "?" + ts.UrlInfo.RawQuery
	} else {
		ts.TsUrl = stream.M3u8UrlInfo.Scheme + "://" + stream.M3u8UrlInfo.Host + path.Dir(stream.M3u8UrlInfo.Path) + "/" + ts.RelativeToParent + "/" + ts.Name
	}

	if stream.LocalPath != "" {
		ts.LocalFile = stream.LocalPath + "/" + ts.RelativeToParent + "/" + ts.Name
	}

	if stream.OnlyM3u8 {
		ts.Status = TsStatusOk
		ts.Size = 1024
	} else {
		retryCount := 0
		for retryCount < stream.RetryCount {
			size, err := stream.doDownloadTs(ts.TsUrl, ts.LocalFile)
			if err == nil {
				ts.Status = TsStatusOk
				ts.Size = size
				break
			}
			retryCount++
			stream.handleError(ErrorCodeTsDownloadRetry, ts, "TsDownload downloading err=%s tsName=%s tsUrl=%s tsLocalFile=%s timeout=%d retryCount=%d",
				err.Error(), ts.Name, ts.TsUrl, ts.LocalFile, stream.Timeout, retryCount)
			if retryCount >= stream.RetryCount {
				break
			}
		}
		if ts.Status != TsStatusOk {
			ts.Status = TsStatusFail
		}
	}

	results <- ts
}

func (stream *Stream) tsExists(tsName string) bool {
	stream.Lock.Lock()
	defer stream.Lock.Unlock()
	_, exists := stream.TsMap[tsName]
	return exists
}

func (stream *Stream) tsAdd(ts *Ts) {
	stream.Lock.Lock()
	defer stream.Lock.Unlock()

	//清理
	if len(stream.TsMap) >= StreamTsCountMax {
		var keys []string
		for key, _ := range stream.TsMap {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		deleteCount := 0
		for _, key := range keys {
			delete(stream.TsMap, key)
			deleteCount++
			if deleteCount >= StreamTsCountReduce {
				break
			}
		}
	}

	stream.TsMap[ts.Name] = ts
}

func (stream *Stream) downloadM3u8Ts(m3u8 *M3u8) {
	tsCount := 0
	results := make(chan *Ts, tsCount)
	for _, ts := range m3u8.TsEntries {
		if !stream.tsExists(ts.Name) {
			stream.tsAdd(ts)
			go stream.downloadTs(ts, results)
			tsCount++
		}
	}

	// waiting ts download finish
	finish := 0
	for finish < tsCount {
		select {
		case ts := <-results:
			finish++
			if ts.Status != TsStatusOk {
				stream.handleError(ErrorCodeTsDownloadFail, ts, "TsDownload fail err=ts downlad fail! tsName=%s retryCount=%d",
					ts.Name, stream.RetryCount)
			}
		}
		if finish >= tsCount {
			break
		}
	}

	for _, ts := range m3u8.TsEntries {
		if ts.Status == TsStatusOk {
			stream.handleDownloadedTs(ts)
		}
	}
}

func (stream *Stream) Pull() {
	go stream.pull()
}

func (stream *Stream) Stop() {
	stream.Stoped = true
}

func (stream *Stream) StopAndWait() {
	for _, secondStream := range stream.Streams {
		secondStream.StopAndWait()
	}
	stream.Stop()
	if stream.Closed != nil {
		<-stream.Closed
	}
}

func (stream *Stream) DownloadM3u8() (m3u8 *M3u8, err error) {
	m3u8String := ""
	if stream.AntileechRemote != "" {
		headers := http.Header{}
		headers.Set("Strm-Uri", stream.M3u8Url)
		m3u8String, err = httputils.HttpGet(antileech.AntileechUrl(stream.AntileechRemote), stream.Timeout, headers)
	} else {
		m3u8String, err = httputils.HttpGet(stream.M3u8Url, stream.Timeout, nil)
	}
	if err != nil {
		return nil, err
	}

	m3u8 = NewM3u8(stream.M3u8Url)
	m3u8.Parse(m3u8String)

	if !m3u8.IsSecond() && !m3u8.IsTop() {
		return nil, fmt.Errorf("invalid m3u8 format!")
	}

	//save file
	if stream.SaveM3u8 && stream.LocalFile != "" {
		SaveFile(m3u8String, stream.LocalFile)
		m3u8.LocalFile = stream.LocalFile
	}

	return m3u8, nil
}

func (stream *Stream) pullM3u8() {
	m3u8, err := stream.DownloadM3u8()
	if err != nil {
		stream.handleError(ErrorCodeM3u8DownloadFail, nil, "M3u8Download err=%s", err.Error())
		return
	}
	//判断内容是否相同
	if m3u8.M3u8String == stream.M3u8String {
		return
	}

	if !m3u8.IsTop() && !m3u8.IsSecond() {
		stream.handleError(ErrorCodeM3u8FormatError, m3u8, "M3u8Format err=unknown m3u8 format!")
		return
	}
	stream.LastSequence = m3u8.Sequence
	stream.M3u8String = m3u8.M3u8String
	stream.handleParseM3u8(m3u8)

	if m3u8.IsTop() {
		stream.IsTop = true
		urlDir := stream.M3u8UrlInfo.Scheme + "://" + stream.M3u8UrlInfo.Host + path.Dir(stream.M3u8UrlInfo.Path)
		for _, entry := range m3u8.M3u8Entries {
			m3u8Url := urlDir + "/" + entry.RelativeToParent + "/" + entry.Name + "?" + entry.UrlInfo.RawQuery
			secondStream := stream.FindStream(entry.Name)
			if secondStream != nil {
				continue
			}
			secondStream = stream.AddStream(m3u8Url)
			secondStream.Pull()
		}

	} else if m3u8.IsSecond() {
		stream.downloadM3u8Ts(m3u8)
		stream.handleDownloadedM3u8(m3u8)
	}
}

func (stream *Stream) pull() {
	stream.Closed = make(chan struct{}, 1)
	for {
		if stream.Stoped {
			break
		}

		go stream.pullM3u8()
		time.Sleep(time.Second)
	}
	close(stream.Closed)
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
