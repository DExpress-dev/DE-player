package mysql

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type FileInfo struct {
	FileId       int64     "Id"
	Uploader     int64     "uploader"
	Owner        int64     "owner"
	Name         string    "name"
	Ext          string    "ext"
	Size         int64     "size"
	Sha          string    "sha"
	Title        string    "title"
	Tag          string    "tag"
	Description  string    "description"
	IpfsHash     string    "ipfsHash"
	IpfsUrl      string    "ipfsUrl"
	AliDnaJobId  string    "aliDnaJobId"
	AliDnaFileId string    "aliDnaFileId"
	Status       int       "status"
	CreateTime   time.Time "createTime"
}

type UserFile struct {
	Id         int64     "Id"
	FileId     int64     "fileId"
	Uploader   int64     "uploader"
	CreateTime time.Time "createTime"
}

func Test(t *testing.T) {
	err := Initialize("172.200.2.195", "3306", "root", "123456", "bokerchain")
	if err != nil {
		fmt.Println(err)
		return
	}

	//	fmt.Println(UpdateTableWhere("bc_file_info", "Id", 2, "IpfsHash", "hehehe2", "title", "testTitle1"))
	//	fmt.Println(InsertTable("bc_user_file", "fileId", 77, "uploader", 88, "createTime", time.Now().Unix()))
	//	res, err := Query("select * from bc_file_info where Id=1")
	//	fmt.Println(QueryObject(fInfo, "select * from bc_file_info where Id=1"))
	//	fmt.Printf("%#v\n", fInfo)
	//	fmt.Println(QueryObjectEx(reflect.TypeOf(FileInfo{}), "select * from bc_file_info where Id=1"))

	_, err = InsertObject("bc_file_info", &FileInfo{CreateTime: time.Now()})
	if err != nil {
		fmt.Println(err)
		return
	}

	objects, err := QueryObjects(reflect.TypeOf(FileInfo{}), "select * from bc_file_info where Id=?", 1)
	for _, object := range objects {
		fmt.Printf("%#v\n", object)
		fInfo := object.(*FileInfo)
		fmt.Println(fInfo.CreateTime)
	}
}
