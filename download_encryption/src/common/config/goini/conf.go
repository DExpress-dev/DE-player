/**
 * Read And Write the configuration file
 *
 * @copyright           (C) 2017  fxh7622
 * @lastmodify          2017-9-26
 * @website		http://www.widuu.com
 *
 */

package goini

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

//key信息
type keyContext struct {
	isAnnotation bool   //是否是注释
	key          string //Key
	context      string //内容
}

//session信息
type sessionContext struct {
	isAnnotation bool       //是否是注释
	session      string     //session值
	keyArray     *list.List //key是key
}

type Config struct {

	//配置文件的路径
	iniPath string

	//内容数据
	sessionContextMap map[string]*sessionContext //key是session
}

//检测错误
func CheckErr(err error) string {
	if err != nil {
		return fmt.Sprintf("Error is :'%s'", err.Error())
	}
	return "Notfound this error"
}

//判断是否是注释
func (configPtr *Config) isAnnotation(line string) bool {
	if string(line[0]) == "#" {
		return true
	} else {
		return false
	}
}

//判断是否是Session
func (configPtr *Config) isSession(line string) (string, bool) {

	if line[0] == '[' && line[len(line)-1] == ']' {
		return strings.TrimSpace(line[1 : len(line)-1]), true
	} else {
		return "", false
	}
}

//判断是否是Key
func (configPtr *Config) isKey(line string) (string, string, bool) {

	pos := strings.IndexAny(line, "=")
	if pos != -1 {

		key := strings.TrimSpace(line[0:pos])
		value := strings.TrimSpace(line[pos+1 : len(line)])

		return key, value, true
	} else {
		return "", "", false
	}
}

//添加注释
func (configPtr *Config) addAnnotation(session, annotation string) {

	if "" == session {

		var sessionPtr *sessionContext = new(sessionContext)
		sessionPtr.isAnnotation = true
		sessionPtr.session = annotation
		sessionPtr.keyArray = list.New()

		configPtr.sessionContextMap[session] = sessionPtr

	} else {
		configPtr.addKeyAnnotation(session, annotation)
	}
}

//添加Session
func (configPtr *Config) addSession(session string) {

	_, result := configPtr.findSession(session)
	if !result {

		var sessionPtr *sessionContext = new(sessionContext)
		sessionPtr.isAnnotation = false
		sessionPtr.session = session
		sessionPtr.keyArray = list.New()

		configPtr.sessionContextMap[session] = sessionPtr
	}
}

//添加Key
func (configPtr *Config) addKey(session, key, context string) {

	sessionContextPtr, result := configPtr.findSession(session)

	if !result {

		var keyPtr *keyContext = new(keyContext)
		keyPtr.isAnnotation = false
		keyPtr.key = key
		keyPtr.context = context

		var sessionContextPtr *sessionContext = new(sessionContext)
		sessionContextPtr.isAnnotation = false
		sessionContextPtr.session = session
		sessionContextPtr.keyArray = list.New()
		sessionContextPtr.keyArray.PushBack(keyPtr)

		configPtr.sessionContextMap[session] = sessionContextPtr
	} else {

		keyPtr, keyResult := configPtr.findKey(session, key)
		if !keyResult {

			var keyPtr *keyContext = new(keyContext)
			keyPtr.isAnnotation = false
			keyPtr.key = key
			keyPtr.context = context

			sessionContextPtr.keyArray.PushBack(keyPtr)
		} else {
			keyPtr.context = context
		}
	}
}

//添加Key的注释
func (configPtr *Config) addKeyAnnotation(session, annotation string) {

	sessionContextPtr, result := configPtr.findSession(session)
	if !result {

		var keyPtr *keyContext = new(keyContext)
		keyPtr.isAnnotation = true
		keyPtr.key = annotation
		keyPtr.context = annotation

		var sessionContextPtr *sessionContext = new(sessionContext)
		sessionContextPtr.isAnnotation = false
		sessionContextPtr.session = session
		sessionContextPtr.keyArray = list.New()
		sessionContextPtr.keyArray.PushBack(keyPtr)

		configPtr.sessionContextMap[session] = sessionContextPtr
	} else {

		keyPtr, keyResult := configPtr.findKey(session, annotation)
		if keyResult {
			keyPtr.context = annotation
		} else {

			var keyPtr *keyContext = new(keyContext)
			keyPtr.isAnnotation = true
			keyPtr.key = annotation
			keyPtr.context = annotation

			sessionContextPtr.keyArray.PushBack(keyPtr)
		}
	}
}

//读取所有Session
func (configPtr *Config) ReadSessions() []string {

	sessionArray := []string{}

	for key, _ := range configPtr.sessionContextMap {
		sessionArray = append(sessionArray, key)
	}
	return sessionArray
}

//读取指定Session下的所有Key
func (configPtr *Config) ReadKeys(session string) []string {

	sessionArray := []string{}
	sessionContextPtr, result := configPtr.findSession(session)
	if !result {
		return sessionArray
	} else {

		for e := sessionContextPtr.keyArray.Front(); e != nil; e = e.Next() {

			keyPtr, ok := e.Value.(*keyContext)
			if ok {
				if !keyPtr.isAnnotation {
					sessionArray = append(sessionArray, keyPtr.key)
				}
			}
		}
		return sessionArray
	}
}

//读取指定Session下Key的内容 string类型
func (configPtr *Config) Read_string(session, key, defaultValue string) string {

	_, result := configPtr.findSession(session)
	if !result {
		return defaultValue
	} else {

		keyPtr, keyResult := configPtr.findKey(session, key)
		if !keyResult {
			return defaultValue
		}
		return keyPtr.context
	}
}

//读取指定Session下Key的内容 int类型
func (configPtr *Config) ReadInt(session, key string, defaultValue int) int {

	_, result := configPtr.findSession(session)
	if !result {
		return defaultValue
	} else {
		keyPtr, keyResult := configPtr.findKey(session, key)
		if !keyResult {
			return defaultValue
		}
		value, _ := strconv.Atoi(keyPtr.context)
		return value
	}
}

//读取指定Session下Key的内容 bool类型
func (configPtr *Config) ReadBool(session, key string, defaultValue bool) bool {

	_, result := configPtr.findSession(session)
	if !result {
		return defaultValue
	} else {
		keyPtr, keyResult := configPtr.findKey(session, key)
		if !keyResult {
			return defaultValue
		}
		value, _ := strconv.Atoi(keyPtr.context)
		if value > 0 {
			return true
		} else {
			return false
		}
	}
}

//写指定Session下Key的内容 string类型
func (configPtr *Config) WriteString(session, key, context string) bool {

	configPtr.addKey(session, key, context)
	configPtr.writeIniFile()
	return true
}

//写指定Session下Key的内容 int类型
func (configPtr *Config) WriteInt(session, key string, context int) bool {

	value := strconv.Itoa(context)
	configPtr.addKey(session, key, value)
	configPtr.writeIniFile()
	return true
}

//写指定Session下Key的内容 bool类型
func (configPtr *Config) WriteBool(session, key string, context bool) bool {

	contextString := ""
	if context {
		contextString = "1"
	} else {
		contextString = "0"
	}
	configPtr.addKey(session, key, contextString)
	configPtr.writeIniFile()
	return true
}

//删除指定Session
func (configPtr *Config) DeleteSession(session string) bool {

	_, result := configPtr.findSession(session)
	if result {
		delete(configPtr.sessionContextMap, session)
		configPtr.writeIniFile()
		return true
	} else {
		return false
	}
}

//填充Ini的Map信息
func (configPtr *Config) fillIniMap(lineArray []string) {

	currentSession := ""

	//轮询填充数据
	for index := range lineArray {

		lineString := lineArray[index]

		session := ""
		key := ""
		context := ""

		//判断是否为空行
		if lineString == "" {
			if currentSession != "" {
				configPtr.addKey(currentSession, key, context)
			} else {
				configPtr.addAnnotation(currentSession, "")
			}
			continue
		}

		//判断是否是注释
		if configPtr.isAnnotation(lineString) {
			configPtr.addAnnotation(currentSession, lineString)
			continue
		}

		session, sessionResult := configPtr.isSession(lineString)
		if sessionResult {
			configPtr.addSession(session)
			currentSession = session
			continue
		}

		key, context, keyResult := configPtr.isKey(lineString)
		if keyResult {
			configPtr.addKey(currentSession, key, context)
		}
	}
}

//查找Session
func (configPtr *Config) findSession(session string) (*sessionContext, bool) {

	sessionContextPtr, exist := configPtr.sessionContextMap[session]
	return sessionContextPtr, exist
}

//查找Key
func (configPtr *Config) findKey(session, key string) (*keyContext, bool) {

	sessionContextPtr, exist := configPtr.sessionContextMap[session]
	if exist {

		for e := sessionContextPtr.keyArray.Front(); e != nil; e = e.Next() {

			keyPtr, ok := e.Value.(*keyContext)
			if ok && keyPtr.key == key {
				return keyPtr, true
			}
		}
	}

	return nil, false
}

//回写Ini文件
func (configPtr *Config) writeIniFile() {

	var iniFileContext string

	//遍历得到所有的
	for _, value := range configPtr.sessionContextMap {

		if value.isAnnotation {

			iniFileContext = iniFileContext + value.session + "\n"

		} else {

			//写session
			iniFileContext = iniFileContext + "[" + value.session + "]\n"

			//写key
			for e := value.keyArray.Front(); e != nil; e = e.Next() {

				keyPtr, keyOk := e.Value.(*keyContext)
				if keyOk {

					if keyPtr.isAnnotation {
						iniFileContext = iniFileContext + keyPtr.key + "\n"
					} else {
						if keyPtr.key == "" {
							iniFileContext = iniFileContext + "\n"
						} else {
							iniFileContext = iniFileContext + keyPtr.key + "=" + keyPtr.context + "\n"
						}

					}
				}
			}
		}
	}

	//写文件
	iniFile, err := os.Create(configPtr.iniPath)

	if err != nil {
		fmt.Println(iniFile, err)
		return
	}
	defer iniFile.Close()

	iniFile.WriteString(iniFileContext)
}

//读取Ini文件
func (configPtr *Config) readIni() {

	//判断ini文件是否存在
	file, err := os.Open(configPtr.iniPath)
	if err != nil {
		CheckErr(err)
	}
	defer file.Close()

	//将所有的line信息放入到一个list中
	lineArray := []string{}

	//读取文件
	buf := bufio.NewReader(file)
	for {
		l, err := buf.ReadString('\n')
		line := strings.TrimSpace(l)
		if err != nil {
			if err != io.EOF {
				CheckErr(err)
			}
			if len(line) == 0 {
				break
			}
		}
		lineArray = append(lineArray, line)
	}

	configPtr.fillIniMap(lineArray)
}

//初始化Ini
func Init(iniPath string) *Config {

	//设置ini文件信息
	configPtr := new(Config)
	configPtr.iniPath = iniPath
	configPtr.sessionContextMap = make(map[string]*sessionContext)

	configPtr.readIni()
	return configPtr
}
