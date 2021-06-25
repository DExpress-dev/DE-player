package hls

import (
	"common/httputils"
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	fmt.Println("Test")
	m3u8Url := "http://pc.oneshow.xyz:81/214/tvonip/live/tempfs/5233/index.m3u8"
	m3u8 := NewM3u8(m3u8Url)
	m3u8String, err := httputils.HttpGet(m3u8Url, 5, nil)
	if err != nil {
		fmt.Println(err)
	}
	m3u8.Parse(m3u8String)
	for _, ts := range m3u8.TsEntries {
		fmt.Printf("%#v\n", ts)
	}
}
