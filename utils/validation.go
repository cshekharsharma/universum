package utils

import (
	"reflect"
)

func IsWriteableDatatype(value interface{}) bool {
	switch value.(type) {
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true

	case []string, []bool,
		[]int, []int8, []int16, []int32, []int64,
		[]uint, []uint8, []uint16, []uint32, []uint64,
		[]float32, []float64:
		return true

	default:
		return false
	}
}

func IsWriteableDataSize(value interface{}, maxsize int64) bool {
	if reflect.ValueOf(value).Kind() == reflect.String {
		if len(value.(string)) > int(maxsize) {
			return false
		}
	}
	return true
}
