package resp3

import (
	"fmt"
	"reflect"
	"strconv"
	"universum/engine/entity"
)

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

	case *entity.Record:
		if v == nil {
			return "_\r\n", nil
		}
		generic := make(map[string]interface{})
		generic["Type"] = v.Type
		generic["LAT"] = v.LAT
		generic["Value"] = v.Value

		return Encode(generic)

	default:
		return "", fmt.Errorf("unsupported type: %v", reflect.TypeOf(value))
	}
}
