package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
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
	var mapHeader int64 = 48 // Overhead for a map structure

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

	case reflect.Map:
		totalSize := mapHeader // Base map header size
		keys := val.MapKeys()

		for _, key := range keys {
			keySize, err := GetSizeInBytes(key.Interface())
			if err != nil {
				return 0, err
			}

			value := val.MapIndex(key).Interface()
			valueSize, err := GetSizeInBytes(value)
			if err != nil {
				return 0, err
			}

			totalSize += keySize + valueSize
		}
		return totalSize, nil

	case reflect.Ptr:
		if val.IsNil() {
			return 0, nil // Size is 0 for nil pointers
		}
		// Get the size of the pointer itself
		ptrSize := int64(unsafe.Sizeof(v))

		// Dereference the pointer and get the size of the underlying value
		elemSize, err := GetSizeInBytes(val.Elem().Interface())
		if err != nil {
			return 0, err
		}
		return ptrSize + elemSize, nil

	case reflect.Struct:
		var structSize int64
		structSize = int64(val.Type().Size()) // Account for struct size, including padding
		return structSize, nil

	default:
		return 0, fmt.Errorf("unsupported type: %s", val.Kind())
	}
}
