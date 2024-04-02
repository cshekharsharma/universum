package engine

import (
	"universum/utils"
)

const (
	TYPE_ENCODING_INT    uint8 = 1
	TYPE_ENCODING_FLOAT  uint  = 2
	TYPE_ENCODING_STRING uint8 = 3
	TYPE_ENCODING_ARRAY  uint8 = 4
	TYPE_ENCODING_MAP    uint8 = 5
)

func GetTypeEncoding(v any) uint8 {
	oType := TYPE_ENCODING_STRING

	if utils.IsInteger(v) {
		return TYPE_ENCODING_INT
	}

	return oType
}
