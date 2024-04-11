package utils

import "time"

func GetCurrentEPochTime() uint32 {
	return uint32(time.Now().Unix())
}

func SecondsSince(since uint32) uint32 {
	return GetCurrentEPochTime() - since
}
