package redis

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

var _pool *redis.Pool = nil

func Initialize(ip, port, password string, maxIdle, maxActive int) (err error) {
	_pool = &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			var c redis.Conn = nil
			if "" == password {
				c, err = redis.Dial("tcp", fmt.Sprintf("%s:%s", ip, port))
			} else {
				c, err = redis.Dial("tcp", fmt.Sprintf("%s:%s", ip, port), redis.DialPassword(password))
			}
			if err != nil {
				return nil, err
			}
			//			c.Do("SELECT", REDIS_DB)
			return c, nil
		},
	}

	conn := _pool.Get()
	if nil == conn {
		return fmt.Errorf("redis not connected!")
	}
	defer conn.Close()

	_, err = conn.Do("ping")
	if err != nil {
		return err
	}

	return nil
}

type Result struct {
	reply interface{}
}

func NewResult() *Result {
	res := &Result{}
	return res
}

func (res *Result) IsNil() bool {
	return nil == res.reply
}

func (res *Result) GetString() (value string, err error) {
	value, err = redis.String(res.reply, nil)
	if err != nil {
		return "", fmt.Errorf("redis Result.GetString err:%s", err.Error())
	}
	return value, nil
}

func (res *Result) GetInt() (value int, err error) {
	value, err = redis.Int(res.reply, nil)
	if err != nil {
		return -1, fmt.Errorf("redis Result.GetInt err:%s", err.Error())
	}
	return value, nil
}

func (res *Result) GetInt64() (value int64, err error) {
	value, err = redis.Int64(res.reply, nil)
	if err != nil {
		return -1, fmt.Errorf("redis Result.GetInt64 err:%s", err.Error())
	}
	return value, nil
}

func (res *Result) GetStringArray() (strs []string, err error) {
	strs, err = redis.Strings(res.reply, nil)
	if err != nil {
		return nil, fmt.Errorf("redis Result.GetStringArray err:%s", err.Error())
	}
	return strs, nil
}

func (res *Result) GetStringMap() (strMap map[string]string, err error) {
	strMap, err = redis.StringMap(res.reply, nil)
	if err != nil {
		return nil, fmt.Errorf("redis Result.GetStringMap err:%s", err.Error())
	}
	return strMap, nil
}

func Exec(commandName string, args ...interface{}) (res *Result, err error) {
	if nil == _pool {
		return nil, fmt.Errorf("redis not initialized!")
	}

	conn := _pool.Get()
	if nil == conn {
		return nil, fmt.Errorf("redis not connected!")
	}
	defer conn.Close()

	reply, err := conn.Do(commandName, args...)
	if err != nil {
		return nil, err
	}
	res = NewResult()
	res.reply = reply

	return res, nil
}

func Del(key string) (err error) {
	_, err = Exec("DEL", key, nil)
	return err
}

func redisValue(value interface{}) interface{} {
	if valueTime, ok := value.(time.Time); ok {
		return valueTime.Unix()
	} else {
		return value
	}
	return nil
}

func Set(key, value interface{}) (err error) {
	_, err = Exec("SET", key, redisValue(value))
	return err
}

func SetEx(key, value interface{}, seconds int64) (err error) {
	_, err = Exec("SETEX", key, seconds, redisValue(value))
	return err
}

func Get(key string) (res *Result, err error) {
	res, err = Exec("GET", key)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func HMSet(key string, args ...interface{}) (err error) {
	argLen := len(args)
	if argLen <= 0 {
		return fmt.Errorf("args length less equal 0")
	}

	if (argLen)%2 != 0 {
		return fmt.Errorf("args not in pair")
	}
	var args2 []interface{}
	args2 = append(args2, key)
	for _, arg := range args {
		args2 = append(args2, redisValue(arg))
	}

	_, err = Exec("HMSET", args2...)
	return err
}

func HGet(key, field string) (res *Result, err error) {
	res, err = Exec("HGET", key, field)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func HGetAll(key string) (res *Result, err error) {
	res, err = Exec("HGETALL", key)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func HGetAllObject(object interface{}, key string) error {
	//check object
	objectValue := reflect.ValueOf(object)
	if objectValue.Kind() != reflect.Ptr || objectValue.IsNil() {
		return fmt.Errorf("object should be pointer to struct")
	}

	elemValue := objectValue.Elem()
	if elemValue.Kind() != reflect.Struct {
		return fmt.Errorf("object should be pointer to struct")
	}

	res, err := HGetAll(key)
	if err != nil {
		return err
	}
	values, err := res.GetStringMap()
	if err != nil {
		return err
	}

	elemType := elemValue.Type()
	for j := 0; j < elemValue.NumField(); j++ {
		field := elemValue.Field(j)
		fieldType := elemType.Field(j)
		fieldTag := string(fieldType.Tag)
		fieldString, exist := values[fieldTag]
		if !exist {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(fieldString)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			num, err := strconv.ParseInt(fieldString, 10, 64)
			if err != nil || field.OverflowInt(num) {
				continue
			}
			field.SetInt(num)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			num, err := strconv.ParseUint(fieldString, 10, 64)
			if err != nil || field.OverflowUint(num) {
				continue
			}
			field.SetUint(num)
		case reflect.Float32, reflect.Float64:
			num, err := strconv.ParseFloat(fieldString, field.Type().Bits())
			if err != nil || field.OverflowFloat(num) {
				continue
			}
			field.SetFloat(num)
		case reflect.Struct:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				unix, err := strconv.ParseInt(fieldString, 10, 64)
				if err != nil {
					continue
				}
				field.Set(reflect.ValueOf(time.Unix(unix, 0)))
			}
		default:
			return fmt.Errorf("unsupoorted type!")
			//	case reflect.Bool:
			//	case reflect.Complex64, reflect.Complex128:
			//	case reflect.Array, reflect.Slice:
			//	case reflect.Struct
		}
	}

	return nil
}

func HDEL(key, field interface{}) (err error) {
	_, err = Exec("HDEL", key, field)
	return err
}

func SAdd(key string, args ...interface{}) (err error) {
	argLen := len(args)
	if argLen <= 0 {
		return fmt.Errorf("args length less equal 0")
	}

	var args2 []interface{}
	args2 = append(args2, key)
	for _, arg := range args {
		args2 = append(args2, redisValue(arg))
	}

	_, err = Exec("SADD", args2...)
	return err
}

func SIsMember(key string, member interface{}) (is bool, err error) {
	res, err := Exec("SISMEMBER", key, member)
	if err != nil {
		return false, err
	}
	intVal, err := res.GetInt()
	if err != nil {
		return false, err
	}
	is = (intVal == 1)
	return is, nil
}

func SMembers(key string) (res *Result, err error) {
	res, err = Exec("SMEMBERS", key)
	if err != nil {
		return nil, err
	}
	return res, nil
}
