package mysql

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var defaultClient *MysqlClient

func Initialize(ip, port, user, password, db string) (err error) {
	defaultClient, err = NewMysqlClient(ip, port, user, password, db)
	if err != nil {
		return err
	}
	return nil
}

type Entry map[string]interface{}

func (entry Entry) HasColumn(key string) bool {
	_, exist := entry[key]
	return exist
}

func (entry Entry) GetString(key string) (value string) {
	if v, ok := entry[key]; ok {
		if str, ok := v.(string); ok {
			value = str
		}
	}
	return value
}

func (entry Entry) GetBool(key string) (value bool) {
	if v, ok := entry[key]; ok {
		if str, ok := v.(string); ok {
			i, err := strconv.Atoi(str)
			if err != nil {
				return false
			}
			if i != 0 {
				return true
			}
		}
	}
	return false
}

func (entry Entry) GetInt(key string) (value int) {
	if v, ok := entry[key]; ok {
		if str, ok := v.(string); ok {
			i, err := strconv.Atoi(str)
			if err != nil {
				return 0
			}
			return i
		}
	}
	return 0
}

func (entry Entry) GetInt64(key string) (value int64) {
	if v, ok := entry[key]; ok {
		if str, ok := v.(string); ok {
			i, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return 0
			}
			return i
		}
	}
	return 0
}

func (entry Entry) GetUInt64(key string) (value uint64) {
	if v, ok := entry[key]; ok {
		if str, ok := v.(string); ok {
			i, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				return 0
			}
			return i
		}
	}
	return 0
}

func (entry Entry) GetFloat64(key string, bitSize int) (value float64) {
	if v, ok := entry[key]; ok {
		if str, ok := v.(string); ok {
			i, err := strconv.ParseFloat(str, bitSize)
			if err != nil {
				return 0
			}
			return i
		}
	}
	return 0
}

func (entry Entry) GetTimeUnix(key string) (value int64) {
	if v, ok := entry[key]; ok {
		if str, ok := v.(string); ok {
			value, _ := time.ParseInLocation("2006-01-02 15:04:05", str, time.UTC)
			return value.Unix()
		}
	}
	return 0
}

func (entry Entry) GetUTCTime(key string) (value time.Time) {
	if v, ok := entry[key]; ok {
		if str, ok := v.(string); ok {
			value, _ := time.ParseInLocation("2006-01-02 15:04:05", str, time.UTC)
			return value
		}
	}
	return time.Unix(0, 0)
}

func (entry Entry) GetLocalTime(key string) (value time.Time) {
	if v, ok := entry[key]; ok {
		if str, ok := v.(string); ok {
			value, _ := time.ParseInLocation("2006-01-02 15:04:05", str, time.UTC)
			return value.Local()
		}
	}
	return time.Unix(0, 0)
}

type Result struct {
	LastInsertId int64
	RowsAffected int64
	Entries      []Entry
}

func NewResult() *Result {
	res := &Result{
		Entries: make([]Entry, 0),
	}
	return res
}

func FieldSet(field *reflect.Value, fieldType *reflect.StructField, entry *Entry) error {
	fieldTag := string(fieldType.Tag)
	switch field.Kind() {
	case reflect.String:
		field.SetString(entry.GetString(fieldTag))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num := entry.GetInt64(fieldTag)
		if field.OverflowInt(num) {
			return fmt.Errorf("field overflow int!")
		}
		field.SetInt(num)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		num := entry.GetUInt64(fieldTag)
		if field.OverflowUint(num) {
			return fmt.Errorf("field overflow uint!")
		}
		field.SetUint(num)
	case reflect.Float32, reflect.Float64:
		num := entry.GetFloat64(fieldTag, field.Type().Bits())
		if field.OverflowFloat(num) {
			return fmt.Errorf("field overflow float!")
		}
		field.SetFloat(num)
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			field.Set(reflect.ValueOf(entry.GetLocalTime(fieldTag)))
		}
	default:
		return fmt.Errorf("unsupoorted type!")
		//	case reflect.Bool:
		//	case reflect.Complex64, reflect.Complex128:
		//	case reflect.Array, reflect.Slice:
		//	case reflect.Struct
	}
	return nil
}

func Exec(query string, args ...interface{}) (*Result, error) {
	if nil == defaultClient {
		return nil, fmt.Errorf("mysql not connected!")
	}

	return defaultClient.Exec(query, args...)
}

func Query(query string, args ...interface{}) (*Result, error) {
	if nil == defaultClient {
		return nil, fmt.Errorf("mysql not connected!")
	}

	return defaultClient.Query(query, args...)
}

func QueryObject(object interface{}, query string, args ...interface{}) error {
	if nil == defaultClient {
		return fmt.Errorf("mysql not connected!")
	}

	return defaultClient.QueryObject(object, query, args...)
}

func QueryObjectEx(objectType reflect.Type, query string, args ...interface{}) (interface{}, error) {
	if nil == defaultClient {
		return nil, fmt.Errorf("mysql not connected!")
	}

	return defaultClient.QueryObjectEx(objectType, query, args...)
}

func QueryObjects(objectType reflect.Type, query string, args ...interface{}) (objectPtrs []interface{}, err error) {
	if nil == defaultClient {
		return nil, fmt.Errorf("mysql not connected!")
	}

	return defaultClient.QueryObjects(objectType, query, args...)
}

func InsertObject(tableName string, objectPtr interface{}) (*Result, error) {
	if nil == defaultClient {
		return nil, fmt.Errorf("mysql not connected!")
	}

	return defaultClient.InsertObject(tableName, objectPtr)
}

func InsertTable(tableName string, args ...interface{}) (*Result, error) {
	if nil == defaultClient {
		return nil, fmt.Errorf("mysql not connected!")
	}

	return defaultClient.InsertTable(tableName, args...)
}

func UpdateTableWhere(tableName string, whereField string, whereValue interface{}, args ...interface{}) (*Result, error) {
	if nil == defaultClient {
		return nil, fmt.Errorf("mysql not connected!")
	}

	return defaultClient.UpdateTableWhere(tableName, whereField, whereValue, args...)
}
