package utils

import (
	"bufio"
	"fmt"
	"reflect"
	"strconv"
	"universum/engine/entity"
)

func EncodedResponse(response interface{}) string {
	encoded, err := EncodeRESP3(response)

	if err != nil {
		newErr := fmt.Errorf("unexpected error occured while processing output: %v", err)
		encoded, _ = EncodeRESP3(newErr)
	}

	return encoded
}

func EncodeRESP3(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return "$" + strconv.Itoa(len(v)) + "\r\n" + v + "\r\n", nil

	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return ":" + fmt.Sprintf("%d", v) + "\r\n", nil

	case []interface{}:
		resp := "*" + strconv.Itoa(len(v)) + "\r\n"
		for _, elem := range v {
			encodedElem, err := EncodeRESP3(elem)
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
			valueStr, err := EncodeRESP3(vx)
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
		generic["TypeEncoding"] = v.TypeEncoding
		generic["LastAccessedAt"] = v.LastAccessedAt
		generic["Value"] = v.Value

		return EncodeRESP3(generic)

	default:
		return "", fmt.Errorf("unsupported type: %v", reflect.TypeOf(value))
	}
}

// decodeRESP3 decodes a RESP3 encoded message from a bufio.Reader.
// It returns the decoded message as an interface{}, which can be type asserted as needed.
//
// # Sample code for using RESP3 parser
//
//	Sample input string: "*5\r\n$4\r\nMSET\r\n$4\r\nkey1\r\n$16\r\nvalue1 dash dash\r\n$4\r\nkey2\r\n$6\r\nvalue2\r\n"
func DecodeRESP3(reader *bufio.Reader) (interface{}, error) {
	dataType, err := reader.ReadByte()

	if err != nil {
		return nil, err
	}

	switch dataType {
	case '+': // Simple String
		line, _, err := reader.ReadLine()

		if err != nil {
			return nil, err
		}

		return string(line), nil

	case '-': // Error
		line, _, err := reader.ReadLine()

		if err != nil {
			return nil, err
		}

		return fmt.Errorf(string(line)), nil

	case ':': // Integer
		line, _, err := reader.ReadLine()

		if err != nil {
			return nil, err
		}

		return strconv.Atoi(string(line))

	case '$': // Bulk String
		lengthStr, _, err := reader.ReadLine()
		if err != nil {
			return nil, err
		}

		length, err := strconv.Atoi(string(lengthStr))
		if err != nil {
			return nil, err
		}

		if length == -1 {
			return nil, nil // Null bulk string
		}

		value := make([]byte, length)
		_, err = reader.Read(value)
		if err != nil {
			return nil, err
		}
		_, err = reader.Discard(2)

		if err != nil {
			return nil, err
		}
		return string(value), nil

	case '*': // Array
		countStr, _, err := reader.ReadLine()
		if err != nil {
			return nil, err
		}

		count, err := strconv.Atoi(string(countStr))
		if err != nil {
			return nil, err
		}

		if count == -1 {
			return nil, nil // Null array
		}

		array := make([]interface{}, count)

		for i := 0; i < count; i++ {
			element, err := DecodeRESP3(reader)

			if err != nil {
				return nil, err
			}

			array[i] = element
		}
		return array, nil

	case '#':
		b, err := reader.ReadByte()
		if err != nil {
			return false, err
		}

		reader.Discard(2)
		return b == 't', nil

	case '%': // Map of interface{}
		line, _, _ := reader.ReadLine()
		size, _ := strconv.Atoi(string(line))
		resultMap := make(map[string]interface{}, size/2)
		for i := 0; i < size; i += 2 {
			key, _ := DecodeRESP3(reader)
			value, _ := DecodeRESP3(reader)
			resultMap[key.(string)] = value
		}
		return resultMap, nil

	case '_':
		reader.Discard(2) // Discard the trailing \r\n
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported data type: %v", dataType)
	}
}
