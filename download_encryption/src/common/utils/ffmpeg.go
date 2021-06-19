package utils

import (
	"bytes"
	log "common/log4go"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	//_ "github.com/go-sql-driver/mysql"
)

type Transcode struct {
}

func trimstring(src string) string {
	ret := strings.Replace(src, " ", "", -1)
	ret = strings.Replace(ret, "\n", "", -1)
	return ret
}

//get gpu devname
func (trans *Transcode) GetGpuID(gpuNum int, gpuTasks int) string {
	ret := ""
	if gpuNum > 0 {
		t := time.Now()
		sec := int(t.Unix())
		for j := 0; j < gpuNum; j++ {
			var ngpuSeq = sec % gpuNum
			sec++
			sseq := strconv.Itoa(ngpuSeq)
			flag := "hwaccel_device " + sseq
			checkFFmpegCmd := fmt.Sprintf("ps -ef |grep \"%s\" |grep -v grep|wc -l", flag)
			snum := StartCmdAndStdout(checkFFmpegCmd)
			snum = trimstring(snum)
			nnum, _ := strconv.Atoi(snum)
			nnum = nnum / 2
			log.Info("devname:%d,task num:%d", sseq, nnum)
			if nnum < gpuTasks {
				ret = sseq
				log.Info("ret gpuid :%d", ret)
				break
			}
		}
	}
	return ret
}

func (trans *Transcode) CheckAvailFile(inputFile, outputFile, hlsType, hlsTime, isTrans string) (bool, error) {
	errstring := ""
	if b, _ := PathExist(outputFile); !b {
		errstring = "Transcode output file not exist"
		log.Error("in CheckAvailFile,%s", errstring)
		return false, errors.New(errstring)
	}

	if hlsType == "1" {
		byteSlice, err := ioutil.ReadFile(outputFile)
		if err != nil {
			errstring = "read file error"
			log.Error("in CheckAvailFile,%s", errstring)
			return false, errors.New(errstring)
		}

		bodySlice := strings.Split(string(byteSlice), "\n")
		if !strings.HasPrefix(bodySlice[len(bodySlice)-2], "#EXT-X-ENDLIST") {
			errstring = "endlist not found in Transcode output file"
			log.Error("in CheckAvailFile,%s", errstring)
			return false, errors.New(errstring)
		}

		for _, line := range bodySlice {
			if strings.HasPrefix(line, "#EXTINF") {
				start := strings.IndexByte(line, ':')
				end := strings.IndexByte(line, '.')
				duration := line[start+1 : end]
				realDur, _ := strconv.Atoi(duration)
				wantDur, _ := strconv.Atoi(hlsTime)
				if isTrans == "1" {
					if realDur-wantDur > 2 {
						errstring = "trans:ts time abnormal,max > hls_time+2"
						log.Error("in CheckAvailFile,hlstime:%s,reltime:%s", hlsTime, duration)
						return false, errors.New(errstring)
					}
				} else if isTrans == "0" {
					if realDur-wantDur > 5 {
						errstring = "copy:ts time abnormal,max > hls_time+5"
						log.Error("in CheckAvailFile,hlstime:%s,reltime:%s", hlsTime, duration)
						return false, errors.New(errstring)
					}
				}
			}
		}
	} else if hlsType == "0" {
		inputMillSec := GetDuration(inputFile)
		outputMillSec := GetDuration(outputFile)
		if inputMillSec <= 0 || outputMillSec <= 0 {
			log.Error("in CheckAvailFile,get duration 0,inputMillSec:%d,outputMillSec%d ", inputMillSec, outputMillSec)
		}
		timepsan := inputMillSec - outputMillSec
		log.Debug("in CheckAvailFile,timespan:%d", timepsan)
		if timepsan > 1000.0 {
			tin := strconv.Itoa(inputMillSec)
			tout := strconv.Itoa(outputMillSec)
			errstring = "time delta between input and output file more than 1s," + tin + "-" + tout
			log.Error("in CheckAvailFile,%s", errstring)
			return false, errors.New(errstring)
		}
	}

	return true, nil
}

func (trans *Transcode) GetAudio(acodec, atrack, ac, ar, ab string) (string, error) {
	var cmd string
	//audio track
	if atrack == "0" {
		//atrack orig
		cmd += " -map 0:a"
	} else if atrack == "1" {
		cmd += " -map 0:a:0 "
	} else if atrack == "2" {
		cmd += " -map 0:a:1 "
	} else if atrack == "3" {
		cmd += " -map 0:a:2 "
	}

	if acodec != "" {
		if acodec == "copy" {
			cmd += " -acodec " + acodec
			return cmd, nil
		}
		if strings.Contains(acodec, "aac") {
			cmd += " -acodec libfdk_aac"
		} else {
			cmd += " -acodec " + acodec
		}
	} else {
		cmd += " -acodec  libfdk_aac "
		//cmd += " -acodec  aac "
	}

	// 声道数
	if ac == "left" {
		cmd += " -af \"pan=1c|FC=FL\""
	} else if ac == "right" {
		cmd += " -af \"pan=1c|FC=FR\""
	} else if ac != "" {
		cmd += " -ac " + ac
	}

	// 采样率
	if ar != "" {
		cmd += " -ar " + ar
	}

	// 码率
	if ab != "" {
		cmd += " -ab " + ab
	}
	cmd += " "
	return cmd, nil
}

func getGpuResolution(resolution string) string {
	//like 100*100
	ret := ""
	r := strings.Split(resolution, "x")
	if len(r) > 1 {
		w := r[0]
		h := r[1]
		ret = " -vf scale_npp=" + w + ":" + h + " "
	}
	return ret
}
func (trans *Transcode) GetFixedPre(inputPath string) (string, bool, bool) {
	var bPScan bool = false
	var bRotate bool = false
	ret := ""
	//check src scan mode,bPscan true:progressive ,false:interlace(need add deinterlace)
	bPScan = isProgressScan(inputPath)
	bRotate = isRotate(inputPath)
	return ret, bPScan, bRotate
}

func (trans *Transcode) GetGpuDec(srcVCodec string, bReadyGpuDec bool) (string, bool) {
	ret := ""
	var bUsingGpuDec bool = true
	if bReadyGpuDec {
		if srcVCodec == "mpeg2video" {
			ret += " -hwaccel cuvid -c:v mpeg2_cuvid  "
		} else if srcVCodec == "h264" {
			ret += " -hwaccel cuvid -c:v h264_cuvid "
		} else if srcVCodec == "hevc" {
			ret += " -hwaccel cuvid -c:v hevc_cuvid "
		} else {
			bUsingGpuDec = false
		}
	}

	return ret, bUsingGpuDec
}

// inputPath: 视频的绝对路径
func (trans *Transcode) GetVideo(devId string, bUsingGpuDec bool, format, codec, profile, level, s, aspect,
	vbmod, br, maxrate, bufsize, r, keyint_min, g, bf, bframebias, refs, subq,
	top string) (cmd, inputPath string, err error) {

	if format == "avi" && codec == "hevc" {
		return "", "", errors.New("hevc not supported with avi")
	}
	if format == "flv" && codec == "hevc" {
		return "", "", errors.New("hevc not supported with flv")
	}
	if format == "flv" && codec == "prores" {
		return "", "", errors.New("prores not supported with flv")
	}

	if refs != "" {
		nRefs, _ := strconv.Atoi(refs)
		if nRefs < 0 || nRefs > 16 {
			return "", "", errors.New("refs not in the range from 0 to 16")
		}
	}

	//truecodec invole hardcodec
	var trueCodec string
	if codec != "" {
		if devId != "" {
			if codec == "h264" {
				cmd = " -vcodec h264_nvenc "
				trueCodec = "h264_nvenc"
			} else if codec == "hevc" {
				cmd = " -vcodec hevc_nvenc "
				trueCodec = "hevc_nvenc"
			}
		} else {
			cmd = " -vcodec " + codec
			//cpu  only
			//cmd += " -preset veryfast "
		}
	} else {
		cmd = " -vcodec  h264 "
	}

	//SDHE MODE,use devid to specify gpu
	if !bUsingGpuDec && devId != "" {
		cmd += " -gpu "
		cmd += devId
	}

	if profile != "" {
		if codec == "h264" || codec == "h264_nvenc" {
			if level != "" {
				cmd += " -profile:v " + profile + " -level " + level
			} else {
				cmd += " -profile:v " + profile
			}
		}
	}

	if s != "" {
		if bUsingGpuDec {
			new_s := getGpuResolution(s)
			cmd += new_s
		} else {
			cmd += " -s " + s
		}
	}

	if aspect != "" {
		cmd += " -aspect " + aspect
	}

	if vbmod == "vbr" {
		//vbr default
		if maxrate != "" && bufsize != "" {
			cmd += " -b:v " + br + " -maxrate " + maxrate + " -bufsize " + bufsize
		} else {
			cmd += " -b:v " + br
		}
	} else if vbmod == "cbr" {
		//cbr
		cmd += fmt.Sprintf(" -nal-hrd cbr -b:v %s -minrate %s -maxrate %s -bufsize %s", br, br, br, br)
	} else if vbmod == "abr" {
		//abr
		cmd += " -b:v " + br + " -maxrate " + br + " -bufsize " + br
	} else {
		//vbr default
		log.Debug("vbmod is %s,we reset vbmod to vbr,cmd is %s", vbmod, cmd)
		vbmod = "vbr"
		if maxrate != "" && bufsize != "" {
			cmd += " -b:v " + br + " -maxrate " + maxrate + " -bufsize " + bufsize
		} else {
			cmd += " -b:v " + br
		}
	}

	if r != "" {
		cmd += " -r " + r
	} else {
		cmd += " -r 25 "
	}

	if codec == "hevc" {
		//hevc is special
		cmd += " -x265-params \""
		if g != "" {
			cmd += "keyint=" + g + ":"
		}
		if keyint_min != "" {
			cmd += "min-keyint=" + keyint_min + ":"
			cmd += "no-open-gop=1:no-scenecut=1"
		}
		cmd += "\""
	} else if codec == "prores" || codec == "h261" || codec == "mpeg2video" {
		log.Debug("codec:%s,no need and -g -keyint_min -sc_threshold,just skip.", codec)
	} else {
		if g != "" {
			cmd += " -g " + g
		}

		if keyint_min != "" {
			cmd += " -keyint_min " + keyint_min
			cmd += " -sc_threshold 0 "
		}
	}

	//hevc nenc doesnt support bf params
	if profile != "baseline" && trueCodec != "hevc_nvenc" {
		// 最大B帧数
		if bf != "" {
			cmd += " -bf " + bf
		}

		// 最大连续B帧数
		if bframebias != "" {
			cmd += " -bframebias " + bframebias
		}
	}

	// 参考帧
	if refs != "" {
		cmd += " -refs " + refs
	}

	//子像素优化，设置亚像素估计的复杂度
	if subq != "" {
		cmd += " -subq " + subq
	}

	// 隔行扫描，逐行扫描，0逐行，1隔行
	if top == "1" {
		cmd += " -flags ildct+ilme -top 1 "
	}
	cmd += " "
	return cmd, inputPath, nil
}

func (trans *Transcode) GetHls(hlsTime, hlsListSize string) (string, error) {
	var t, size int
	t, err := strconv.Atoi(hlsTime)
	if err != nil {
		return "", err
	}

	size, err = strconv.Atoi(hlsListSize)
	if err != nil {
		return "", err
	}

	if t < 1 || size < 0 {
		return "", errors.New("hls_time or hls_list_size < 1")
	}

	cmd := fmt.Sprintf("-f hls -hls_init_sequence 1 -hls_time %s -hls_list_size %s -hls_init_replay 1",
		hlsTime, hlsListSize)
	return cmd, nil
}

// offset: 距离边缘的像素值
// 			0, 左上角overlay=offset:offset
// 			1, 右上角overlay=main_w-overlay_w-offset:offset
// 			2, 左下角overlay=offset:main_h-overlay_h-offset
// 			3, 右下角overlay=main_w-overlay_w-offset:main_h-overlay_h-offset
func (trans *Transcode) GetWatermark(url, xCoordinate, yCoordinate, watermarkSize,
	corner string) (cmd, watermarkPath string, err error) {
	if url == "" {
		return "", "", errors.New("watermark url is empty")
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "ftp://") {
		watermarkPath = DoLoadFile(url)
	} else {
		watermarkPath = url
	}
	if watermarkPath == "" {
		return "", "", errors.New("watermark path is empty")
	}
	cmd = fmt.Sprintf(" -vf \"movie=%s, scale=%s [watermark];[in][watermark] ", watermarkPath, watermarkSize)
	switch corner {
	case "0":
		cmd += fmt.Sprintf("overlay=%s:%s [out]\" ", xCoordinate, yCoordinate)
	case "1":
		cmd += fmt.Sprintf("overlay=main_w-overlay_w-%s:%s [out]\" ", xCoordinate, yCoordinate)
	case "2":
		cmd += fmt.Sprintf("overlay=%s:main_w-overlay_w-%s [out]\" ", xCoordinate, yCoordinate)
	case "3":
		cmd += fmt.Sprintf("overlay=main_w-overlay_w-%s:main_h-overlay_h-%s [out]\" ", xCoordinate, yCoordinate)
	default:
		return "", watermarkPath, errors.New("unavailable corner")
	}

	return cmd, watermarkPath, nil
}

func GetResolution(path string) (string, error) {
	cmdStr := fmt.Sprintf("ffprobe %s 2>&1 |grep Stream", path)

	cmd := exec.Command("/bin/sh", "-c", cmdStr)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println(cmdStr, err)
		return "", err
	}

	outStr := out.String()
	re, err := regexp.Compile("^ *Stream #.*, (\\d+)x(\\d+).*")
	if err != nil {
		fmt.Println("compile", err)
		return "", err
	}

	resolution := re.FindStringSubmatch(outStr)
	if resolution != nil {
		return resolution[1] + "x" + resolution[2], nil
	}

	return "", errors.New("not get resolution")
}

func GetDuration(path string) int {
	cmdStr := fmt.Sprintf("ffprobe %s 2>&1 |grep Duration", path)
	log.Debug("getDuration cmd:%s", cmdStr)

	cmd := exec.Command("/bin/sh", "-c", cmdStr)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Error("%s,error is :%s", cmdStr, err.Error())
		return -1
	}

	outStr := out.String()
	re, err := regexp.Compile("^ *Duration: *(.*)\\.(\\d*), start.*")
	if err != nil {
		fmt.Println("compile", err)
		return -1
	}

	duration := re.FindStringSubmatch(outStr)
	if duration != nil {
		// Duration: 00:30:00.02, 转换为ms
		leftMillSec, _ := GetSecond(duration[1])
		leftMillSec *= 1000

		rightMillSec, _ := strconv.Atoi(duration[2])
		rightMillSec *= 10

		return leftMillSec + rightMillSec
	}

	return -1
}

//1:00:2 --->3602秒；
func GetSecond(time string) (int, error) {
	strTime := strings.Split(time, ":")

	if len(strTime) < 3 {
		//return 0, errors.New("time format err")
		return 0, nil
	}

	h, err := strconv.Atoi(strTime[0])
	if err != nil {
		return 0, err
	}
	duration := h * 3600

	m, err := strconv.Atoi(strTime[1])
	if err != nil {
		return 0, err
	}
	duration += m * 60

	s, err := strconv.Atoi(strTime[2])
	if err != nil {
		return 0, err
	}

	duration += s
	return duration, nil
}

func GetProgress(filePath string, totalSec int) (int, bool) {
	lastLineCmd := "cat " + filePath + " | tail -n 1"

	lastLine := StartCmdAndStdout(lastLineCmd)

	//fmt.Println("lastLine", lastLine)
	if lastLine == "" {
		return -1, false
	}

	re, err := regexp.Compile("frame=.*fps=.*time=(.*)\\.\\d+ bitrate=.*")
	if err != nil {
		fmt.Println("compile", err)
		return -1, false
	}

	elapseTime := re.FindStringSubmatch(lastLine)
	if elapseTime != nil {
		fmt.Println(elapseTime[1])
		elapseSec, _ := GetSecond(elapseTime[1])
		progress := elapseSec * 100 / totalSec
		return progress, true
	}

	return -1, false
}

func insertTable(transDB *sql.DB, taskId, serverId, outFmt string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	stmt, err := transDB.Prepare("INSERT trans_status SET cmd=?,taskid=?,result=?,serverid=?,progress=?,starttime=?,updatetime=?,outformat=?")
	if err != nil {
		fmt.Println("prepare INSERT trans_status SET taskid... failed", err)
		return err
	}

	ret, err := stmt.Exec("", taskId, 1, serverId, 0, now, now, outFmt)
	if err != nil {
		fmt.Println("exec INSERT trans_status SET taskid... failed", err)
		return err
	}
	num, _ := ret.RowsAffected()
	if num == 0 {
		fmt.Println("insert trans status no affected")
	}

	fmt.Println("insert trans status ok")
	return nil
}

/*
make timeout to fix some ffprobe/ffmpeg process can't return issue
arg			ffmpeg/ffprobe execute command. like ffprobe -i rtmp://ip:port
timeout	    int, like 20,mean timeout 20s
ret			error and "" or execute cmd return info.
*/
func MakeTimeout(arg string, timeout int) (error, string) {
	ch := make(chan string)
	go func() {
		ffargscmd := exec.Command("/bin/sh", "-c", arg)
		out, err := ffargscmd.Output()
		if err != nil {
			log.Error("%s,%s", arg, err.Error())
			//exec failed
			ch <- "fail"
		} else {
			ch <- string(out)
		}
	}()

	select {
	case recv := <-ch:
		if recv == "fail" {
			return nil, ""
		} else {
			info := string(recv)
			info = trimstring(info)
			return nil, info
		}

	case <-time.After(time.Second * time.Duration(timeout)):
		log.Error("timeout %ds so report error,cmd:%s", timeout, arg)
		return nil, ""
	}
}

func isProgressScan(src string) bool {
	arg := "ffmpeg -y -i  " + src + " -vf showinfo -an  -frames:v 2 -f flv  /dev/null 2>&1"
	log.Debug("in scan progress ,arg:%s", arg)
	_, info := MakeTimeout(arg, 20)
	b := strings.Contains(info, "i:P")
	return b
}

func isRotate(src string) bool {
	arg := "ffprobe " + src + " -show_streams|grep rotate"
	//log.Debug("isRotate arg:%s", cmd)
	_, info := MakeTimeout(arg, 20)
	if info == "" {
		return false
	}
	return true
}

func updateTable(transDB *sql.DB, progress int, updateSql, updateTime, taskId string) error {
	var ret sql.Result
	var err error

	fmt.Println("progress: ", progress)

	ret, err = transDB.Exec(updateSql, progress, updateTime, taskId)

	if err != nil {
		fmt.Println("UpdateTranscodeProgress", err)
	} else {
		num, _ := ret.RowsAffected()
		if num == 0 {
			fmt.Println("UpdateTranscodeProgress no affected")
		}

		fmt.Println("UpdateTranscodeProgress ok")
	}

	return err
}

func StartCmdAndStdout(cmdStr string) string {
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Debug("cmd str:%s,%s", cmdStr, err.Error())
		return ""
	}

	outStr := out.String()

	return outStr
}

func StartCmd(cmdStr string) error {
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	err := cmd.Run()
	if err != nil {
		log.Error(cmdStr, err)
		return err
	}

	return nil
}

func (trans *Transcode) HlsSnap(m3u3path string, outpath string,
	getpicpath string, w string, h string) error {

	dir, _ := GetDirName(m3u3path)
	cmd := "f=\"" + m3u3path + "\";cat $f|grep ts"
	cmdret := exec.Command("/bin/sh", "-c", cmd)
	out, err := cmdret.Output()
	if err != nil {
		fmt.Println(err.Error())
		return errors.New("exec Output faild")
	}
	bodySlice := strings.Split(string(out), "\n")
	num := len(bodySlice)
	fmt.Println("hls_snap", cmd, num)

	if num > 1 {
		for j := 0; j < num-1; j++ {
			index := strconv.Itoa(j)
			cmd = getpicpath + " " + dir + "/" + bodySlice[j] + " " + outpath + "/" + index + ".jpg " + " " + w + " " + h
			fmt.Println("hls_snap", cmd)
			lsCmd := exec.Command("/bin/sh", "-c", cmd)
			err := lsCmd.Run()
			if err != nil {
				return errors.New("exec Run faild" + err.Error())
			}
		}
	}

	return nil
}

func (trans *Transcode) GetGpuNum() int {
	//flag is Tesla,if change card brand.can revise this or use config to define
	checkGpuCmd := "nvidia-smi|grep Tesla|wc -l"

	snum := StartCmdAndStdout(checkGpuCmd)
	snum = trimstring(snum)
	nGpuNum, _ := strconv.Atoi(snum)

	return nGpuNum
}
