package utils

import (
	"fmt"
	"reflect"
	"runtime"
)

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

func GetSizeInBytes(v interface{}) (int64, error) {
	val := reflect.ValueOf(v)
	var stringHeader int64 = 16
	var sliceHeader int64 = 24

	switch val.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return int64(val.Type().Size()), nil

	case reflect.String:
		return stringHeader + int64(len(val.String())), nil

	case reflect.Slice:
		totalSize := sliceHeader

		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i).Interface()

			elemSize, err := GetSizeInBytes(elem)
			if err != nil {
				return 0, err
			}
			totalSize += elemSize
		}
		return totalSize, nil

	default:
		return 0, fmt.Errorf("unsupported type: %s", val.Kind())
	}
}
