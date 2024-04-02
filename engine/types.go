package engine

import (
	"reflect"
	"universum/utils"
)

const (
	TYPE_ENCODING_RAW       uint8 = 0
	TYPE_ENCODING_BOOL      uint8 = 1
	TYPE_ENCODING_INT       uint8 = 2
	TYPE_ENCODING_FLOAT     uint8 = 3
	TYPE_ENCODING_STRING    uint8 = 4
	TYPE_ENCODING_ARRAY     uint8 = 5
	TYPE_ENCODING_SLICE     uint8 = 6
	TYPE_ENCODING_MAP       uint8 = 7
	TYPE_ENCODING_INTERFACE uint8 = 8
)

func GetTypeEncoding(v interface{}) uint8 {

	if reflect.TypeOf(v).Kind() == reflect.Bool {
		return TYPE_ENCODING_BOOL
	}

	if utils.IsInteger(v) {
		return TYPE_ENCODING_INT
	}

	if utils.IsFloat(v) {
		return TYPE_ENCODING_FLOAT
	}

	if reflect.TypeOf(v).Kind() == reflect.String {
		return TYPE_ENCODING_STRING
	}

	if reflect.TypeOf(v).Kind() == reflect.Array {
		return TYPE_ENCODING_ARRAY
	}

	if reflect.TypeOf(v).Kind() == reflect.Slice {
		return TYPE_ENCODING_SLICE
	}

	if reflect.TypeOf(v).Kind() == reflect.Map {
		return TYPE_ENCODING_MAP
	}

	if reflect.TypeOf(v).Kind() == reflect.Interface {
		return TYPE_ENCODING_INTERFACE
	}

	return TYPE_ENCODING_RAW
}
