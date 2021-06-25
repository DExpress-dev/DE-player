package redis

import (
	"fmt"
	"testing"
	"time"
)

func redisTest() {
	Initialize("127.0.0.1", "6379", "", 2, 10)
	key := "mm"
	res, err := Exec("hgetall", key)
	if err != nil {
		fmt.Println("redis get failed:", err)
		return
	}
	if res.IsNil() {
		fmt.Printf("redis %s not found\n", key)
		return
	}

	strs, err := res.GetStringMap()
	if err != nil {
		fmt.Println("redis get failed:", err)
		return
	}
	fmt.Printf("Get mykey: %v \n", strs)
	Exec("HSET", "hash:appuser2account:boker", "1001", "55555")
	res, err = HGet("hash:appuser2account:boker", "1001")
	fmt.Println(res.GetInt())
	SetEx("test", "mayday", 5)
}

type FileInfo struct {
	FileId         int64     "Id"
	Uploader       int64     "uploader"
	CopyrightOwner int64     "copyrightOwner"
	Name           string    "name"
	Ext            string    "ext"
	Size           int64     "size"
	Sha            string    "sha"
	Title          string    "title"
	Tag            string    "tag"
	Description    string    "description"
	IpfsHash       string    "ipfsHash"
	IpfsUrl        string    "ipfsUrl"
	AliDnaJobId    string    "aliDnaJobId"
	AliDnaFileId   string    "aliDnaFileId"
	Status         int       "status"
	CreateTime     time.Time "createTime"
}

func testCommon() {
	res, _ := HGet("hashtest", "a")
	fmt.Println(res.GetInt())
	HMSet("hashtest", "a", 6)
	res, _ = HGet("hashtest", "a")
	fmt.Println(res.GetInt())
	HDEL("hashtest", "a")
	res, _ = HGet("hashtest", "a")
	fmt.Println(res)
}

func Test(t *testing.T) {
	Initialize("127.0.0.1", "6379", "123456", 2, 10)
	testCommon()
}
