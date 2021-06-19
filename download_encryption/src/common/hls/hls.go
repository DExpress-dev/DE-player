// stream_check project hls.go
package hls

import (
	"strconv"
	"strings"
	"time"
)

//m3u8中的TS内容
type TsInfo struct {
	extinf string //时长
	tsPath string //ts路径
}

//m3u8内容
type M3u8Info struct {
	version        int       //版本
	targetduration int       //
	sequence       int       //sequence
	endlist        bool      //是否存在endlist
	tsArray        []*TsInfo //ts信息数组
}

type Hls struct {
}

//得到文件名(不包含路径 /20171226/700/20171226T143444.ts -> 20171226T143444.ts)
func (hlsPtr *Hls) GetRemoteTSName(url string) (string, bool) {

	index := strings.LastIndex(url, "/")
	if index == -1 {
		return "", false
	}
	return string(url[index:len(url)]), true
}

//得到文件路径(不包含文件名 /20171226/700/20171226T143444.ts -> /20171226/700/)
func (hlsPtr *Hls) GetRemoteTSPath(url string) (string, bool) {

	index := strings.LastIndex(url, "/")
	if index == -1 {
		return "", false
	}
	return string(url[0:index]), true
}

//得到本地文件名(700/20180102/20180102T182312.ts -> localpath/700/20180102/20180102T182312.ts)
func (hlsPtr *Hls) GetLocalTsPath(tsName, m3u8Url, tsLocalPath string) (string, bool) {

	return tsLocalPath + tsName, true
}

//得到当前时间 20171226T143444
func (hlsPtr *Hls) GetCurrentTimerISO() string {

	timeNow := time.Now()
	return timeNow.Format("20060102T150405")
}

//将ISO格式转换成Time格式
func (hlsPtr *Hls) ISOToTime(isotime string) (bool, time.Time) {

	timer, err := time.Parse("20060102T150405", isotime)
	if err != nil {
		return false, time.Now()
	}
	return true, timer
}

//拆分函数
func (hlsPtr *Hls) splitM3u8(s rune) bool {
	if s == '\n' {
		return true
	}
	return false
}

//字符串拆分(http://guoguang.live.otvcloud.com/otv/xjgg/live/channel17/700.m3u8)
func (hlsPtr *Hls) m3u8SplitFunc(m3u8Context string) []string {

	return strings.FieldsFunc(m3u8Context, hlsPtr.splitM3u8)
}

//得到version
func (hlsPtr *Hls) getVersion(m3u8Context string) (bool, string) {

	lineArray := hlsPtr.m3u8SplitFunc(m3u8Context)
	for i := range lineArray {

		line := lineArray[i]
		flagIndex := strings.LastIndex(line, "#EXT-X-VERSION:")
		if flagIndex >= 0 {
			return true, line[flagIndex+len("#EXT-X-VERSION:") : len(line)]
		}
	}
	return false, ""
}

//得到targetduration
func (hlsPtr *Hls) getTargetduration(m3u8Context string) (bool, string) {

	lineArray := hlsPtr.m3u8SplitFunc(m3u8Context)
	for i := range lineArray {

		line := lineArray[i]
		flagIndex := strings.LastIndex(line, "#EXT-X-TARGETDURATION:")
		if flagIndex >= 0 {
			return true, line[flagIndex+len("#EXT-X-TARGETDURATION:") : len(line)]
		}
	}
	return false, ""

}

//得到Sequence
func (hlsPtr *Hls) getSequence(m3u8Context string) (bool, string) {

	lineArray := hlsPtr.m3u8SplitFunc(m3u8Context)
	for i := range lineArray {

		line := lineArray[i]
		flagIndex := strings.LastIndex(line, "#EXT-X-MEDIA-SEQUENCE:")
		if flagIndex >= 0 {
			return true, line[flagIndex+len("#EXT-X-MEDIA-SEQUENCE:") : len(line)]
		}
	}
	return false, ""
}

//得到Endlist(此处暂时未完成)
func (hlsPtr *Hls) getEndlist(m3u8Context string) bool {

	lineArray := hlsPtr.m3u8SplitFunc(m3u8Context)
	for i := range lineArray {

		line := lineArray[i]
		flagIndex := strings.LastIndex(line, "#EXT-X-MEDIA-SEQUENCE:")
		if flagIndex >= 0 {
			return true
		}
	}
	return false
}

//得到TS的时间间隔(#EXTINF:2.000000, ->2.000000)
func (hlsPtr *Hls) getTsTimerInterval(line string) (bool, string) {

	//是否是时间;
	flagIndex := strings.LastIndex(line, "#EXTINF:")
	if flagIndex >= 0 {

		commaIndex := strings.LastIndex(line, ",")
		return true, line[flagIndex+len("#EXTINF:") : commaIndex]
	}
	return false, ""
}

//得到Ts数组
func (hlsPtr *Hls) getTsArray(m3u8Context string, m3u8Info *M3u8Info) {

	lineArray := hlsPtr.m3u8SplitFunc(m3u8Context)
	for i := range lineArray {

		line := lineArray[i]

		exist, interval := hlsPtr.getTsTimerInterval(line)
		if exist {

			var tsInfo *TsInfo = new(TsInfo)
			tsInfo.extinf = interval
			tsInfo.tsPath = lineArray[i+1]
			m3u8Info.tsArray = append(m3u8Info.tsArray, tsInfo)
		}

	}
}

//拆分m3u8内容
func (hlsPtr *Hls) SplitM3u8(m3u8Context string) (bool, *M3u8Info) {

	var m3u8Info *M3u8Info = new(M3u8Info)

	//得到Version
	result, resultString := hlsPtr.getVersion(m3u8Context)
	if result {

		ver, err := strconv.Atoi(resultString)
		if err == nil {

			m3u8Info.version = ver
		}
	}

	//得到Targetduration
	result, resultString = hlsPtr.getTargetduration(m3u8Context)
	if result {

		targetduration, err := strconv.Atoi(resultString)
		if err == nil {

			m3u8Info.targetduration = targetduration
		}
	}

	//得到Sequence
	result, resultString = hlsPtr.getSequence(m3u8Context)
	if result {

		sequence, err := strconv.Atoi(resultString)
		if err == nil {

			m3u8Info.sequence = sequence
		}
	}

	//得到Endlist
	m3u8Info.endlist = hlsPtr.getEndlist(m3u8Context)

	//得到Ts数组
	hlsPtr.getTsArray(m3u8Context, m3u8Info)

	//返回
	return true, m3u8Info
}

//Hls管理
func CreateHls() *Hls {

	hlsPtr := new(Hls)
	return hlsPtr
}
