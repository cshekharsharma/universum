package utils

import (
	"time"

	"golang.org/x/exp/rand"
)

func GetRandomString(length int64) string {
	r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}
