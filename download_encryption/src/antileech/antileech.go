package antileech

import (
	"fmt"
	"math/rand"
	"time"
)

func RandomStringByPattern(pattern []byte, n int) string {
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		result = append(result, pattern[r.Intn(len(pattern))])
	}
	return string(result)
}

func RandomString(n int) string {
	pattern := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLOMNOPQRSTUVWXYZ1234567890"
	return RandomStringByPattern([]byte(pattern), n)
}

func AntileechUrl(antileechRemote string) string {
	return fmt.Sprintf("http://%s/key?UID=%s", antileechRemote, RandomString(8))
}
