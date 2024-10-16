package utils

import (
	"encoding/base64"
	"reflect"
	"time"

	"golang.org/x/exp/rand"
)

func IsString(v interface{}) bool {
	return reflect.ValueOf(v).Kind() == reflect.String
}

func GetRandomString(length int64) string {
	r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}

func GetRandomStringCrypto(length uint8) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
