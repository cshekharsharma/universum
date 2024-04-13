package utils

import "time"

func GetCurrentEPochTime() int64 {
	return time.Now().Unix()
}

func SecondsSince(since int64) int64 {
	return GetCurrentEPochTime() - since
}
