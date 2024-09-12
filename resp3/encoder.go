package resp3

import (
	"fmt"
	"reflect"
	"strconv"
	"universum/storage"
)

// Encode converts a Go data type into its corresponding RESP3 encoded string format.
// It supports encoding basic types (string, integers, floats), composite types (slices, maps),
// booleans, nil values, custom error messages, and even specific custom types like
// *storage.ScalarRecord. The function uses type assertion and reflection to determine
// the input type and formats it accordingly.
//
// Parameters:
//   - value interface{}: The value to be encoded into RESP3 format. This could be any supported
//     Go data type, including custom types that have a defined encoding pattern.
//
// Returns:
//   - string: The RESP3 encoded string representation of the input `value`.
//   - error: An error is returned if the value type is not supported for encoding or
//     if any issue arises during the encoding process.
//
// Supported Types:
//   - Basic types: Encodes strings, integers, and floats with respective RESP3 prefixes.
//   - Composite types: Encodes slices as RESP3 arrays and maps as RESP3 maps, recursively encoding
//     their elements.
//   - Booleans: Encodes true and false as RESP3 booleans.
//   - Nil: Encodes a nil value as RESP3 Null.
//   - Errors: Encodes Go error types as RESP3 error messages.
//   - *storage.ScalarRecord: Encodes custom struct types by converting them into a generic map
//     and then encoding this map.
func Encode(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return "$" + strconv.Itoa(len(v)) + "\r\n" + v + "\r\n", nil

	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return ":" + fmt.Sprintf("%d", v) + "\r\n", nil

	case float32, float64:
		return "," + fmt.Sprintf("%f", v) + "\r\n", nil

	case []interface{}:
		resp := "*" + strconv.Itoa(len(v)) + "\r\n"
		for _, elem := range v {
			encodedElem, err := Encode(elem)
			if err != nil {
				return "", err
			}

			resp += encodedElem
		}
		return resp, nil

	case bool:
		if v {
			return "#t\r\n", nil
		}
		return "#f\r\n", nil

	case nil:
		return "_\r\n", nil

	case error:
		return "-" + v.Error() + "\r\n", nil

	case map[string]interface{}:
		resp := "%" + strconv.Itoa(len(v)*2) + "\r\n"
		for kx, vx := range v {
			resp += "+" + kx + "\r\n"
			valueStr, err := Encode(vx)
			if err != nil {
				return "", err
			}
			resp += valueStr
		}
		return resp, nil

	case *storage.ScalarRecord:
		if v == nil {
			return "_\r\n", nil
		}

		scalarRecord := make(map[string]interface{})
		scalarRecord["Type"] = v.Type
		scalarRecord["LAT"] = v.LAT
		scalarRecord["Value"] = v.Value
		scalarRecord["Expiry"] = v.Expiry
		return Encode(scalarRecord)

	case *storage.RecordResponse:
		if v == nil {
			return "_\r\n", nil
		}

		recordResponse := make(map[string]interface{})
		recordResponse["Value"] = v.Value
		recordResponse["Code"] = v.Code
		return Encode(recordResponse)

	default:
		return "", fmt.Errorf("unsupported type: %v", reflect.TypeOf(value))
	}
}
