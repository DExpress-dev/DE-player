package mysql

import (
	"blockchain/BokerChain/common/reflectutils"
	"database/sql"
	"fmt"
	"reflect"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlClient struct {
	MysqlDb *sql.DB
}

func NewMysqlClient(ip, port, user, password, db string) (client *MysqlClient, err error) {
	mysqlParam := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, ip, port, db)
	mysqlDb, err := sql.Open("mysql", mysqlParam)
	if err != nil {
		return nil, err
	}
	err = mysqlDb.Ping()
	if err != nil {
		return nil, err
	}
	return &MysqlClient{MysqlDb: mysqlDb}, nil
}

func (client *MysqlClient) Exec(query string, args ...interface{}) (*Result, error) {
	if nil == client.MysqlDb {
		return nil, fmt.Errorf("mysql not connected!")
	}

	res, err := client.MysqlDb.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	mres := NewResult()

	mres.LastInsertId, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}
	mres.RowsAffected, err = res.RowsAffected()
	if err != nil {
		return nil, err
	}
	return mres, nil
}

func (client *MysqlClient) Query(query string, args ...interface{}) (*Result, error) {
	if nil == client.MysqlDb {
		return nil, fmt.Errorf("mysql not connected!")
	}

	rows, err := client.MysqlDb.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mres := NewResult()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	columnNum := len(columns)
	values := make([]string, columnNum)
	valuePtrs := make([]interface{}, columnNum)
	for i := 0; i < columnNum; i++ {
		valuePtrs[i] = &values[i]
	}
	for rows.Next() {
		rows.Scan(valuePtrs...)
		entry := make(Entry)
		for i, col := range columns {
			value := values[i]
			entry[col] = value
		}
		mres.Entries = append(mres.Entries, entry)
	}
	return mres, nil
}

func (client *MysqlClient) QueryObject(object interface{}, query string, args ...interface{}) error {
	//check object
	objectValue := reflect.ValueOf(object)
	if objectValue.Kind() != reflect.Ptr || objectValue.IsNil() {
		return fmt.Errorf("object should be pointer to struct")
	}

	elemValue := objectValue.Elem()
	if elemValue.Kind() != reflect.Struct {
		return fmt.Errorf("object should be pointer to struct")
	}

	res, err := client.Query(query, args...)
	if err != nil {
		return err
	}
	if len(res.Entries) > 0 {
		elemType := elemValue.Type()
		for j := 0; j < elemValue.NumField(); j++ {
			field := elemValue.Field(j)
			fieldType := elemType.Field(j)
			fieldTag := string(fieldType.Tag)
			if !res.Entries[0].HasColumn(fieldTag) {
				continue
			}

			err = FieldSet(&field, &fieldType, &res.Entries[0])
			if err != nil {
				continue
			}
		}
	}

	return nil
}

func (client *MysqlClient) QueryObjectEx(objectType reflect.Type, query string, args ...interface{}) (interface{}, error) {
	if objectType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("object should be struct")
	}

	res, err := client.Query(query, args...)
	if err != nil {
		return nil, err
	}
	if len(res.Entries) > 0 {
		objectPtr := reflect.New(objectType)
		object := objectPtr.Elem()
		for j := 0; j < object.NumField(); j++ {
			field := object.Field(j)
			fieldType := objectType.Field(j)
			fieldTag := string(fieldType.Tag)
			if !res.Entries[0].HasColumn(fieldTag) {
				continue
			}

			err = FieldSet(&field, &fieldType, &res.Entries[0])
			if err != nil {
				continue
			}
		}
		return objectPtr, nil
	}
	return nil, nil
}

func (client *MysqlClient) QueryObjects(objectType reflect.Type, query string, args ...interface{}) (objectPtrs []interface{}, err error) {
	if objectType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("object should be struct")
	}

	res, err := client.Query(query, args...)
	if err != nil {
		return nil, err
	}

	for _, entry := range res.Entries {
		objectPtr := reflect.New(objectType)
		object := objectPtr.Elem()
		for j := 0; j < object.NumField(); j++ {
			field := object.Field(j)
			fieldType := objectType.Field(j)
			fieldTag := string(fieldType.Tag)
			if !entry.HasColumn(fieldTag) {
				continue
			}

			err = FieldSet(&field, &fieldType, &entry)
			if err != nil {
				continue
			}
		}
		objectPtrs = append(objectPtrs, objectPtr.Interface())
	}
	return objectPtrs, nil
}

func (client *MysqlClient) InsertTable(tableName string, args ...interface{}) (*Result, error) {
	argLen := len(args)
	if argLen <= 0 {
		return nil, fmt.Errorf("args length less equal 0")
	}
	if (argLen)%2 != 0 {
		return nil, fmt.Errorf("args not in pair")
	}

	//fields
	fields := ""
	placeholder := ""
	var values []interface{}
	for i := 0; i < argLen; i += 2 {
		valueName, ok := args[i].(string)
		if !ok {
			return nil, fmt.Errorf("arg invalid")
		}

		if valueName != "" {
			fields += fmt.Sprintf(" `%s`,", valueName)
			placeholder += " ?,"
			values = append(values, args[i+1])
		}
	}

	fields = fields[:len(fields)-1]                //strip last coma
	placeholder = placeholder[:len(placeholder)-1] //strip last coma

	sqlString := fmt.Sprintf("insert into %s (%s) values (%s)", tableName, fields, placeholder)
	return client.Exec(sqlString, values...)
}

func (client *MysqlClient) InsertObject(tableName string, objectPtr interface{}) (*Result, error) {
	helper, err := reflectutils.NewStructHelper(objectPtr)
	if err != nil {
		return nil, fmt.Errorf("reflectutils.NewStructHelper err=%s", err.Error())
	}
	helper.SetAll()
	return client.InsertTable(tableName, helper.SetArgs()...)
}

func (client *MysqlClient) UpdateTableWhere(tableName string, whereField string, whereValue interface{}, args ...interface{}) (*Result, error) {
	argLen := len(args)
	if argLen <= 0 {
		return nil, fmt.Errorf("args length less equal 0")
	}

	if (argLen)%2 != 0 {
		return nil, fmt.Errorf("args not in pair")
	}

	fields := ""
	var values []interface{}
	for i := 0; i < argLen; i += 2 {
		valueName, ok := args[i].(string)
		if !ok {
			return nil, fmt.Errorf("arg invalid")
		}

		if valueName != "" {
			fields += fmt.Sprintf(" `%s`=?,", valueName)
			values = append(values, args[i+1])
		}
	}

	fields = fields[:len(fields)-1]     //strip last coma
	values = append(values, whereValue) //append where value

	sqlString := fmt.Sprintf("update %s set%s where %s=?", tableName, fields, whereField)
	return client.Exec(sqlString, values...)
}
