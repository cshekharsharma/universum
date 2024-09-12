package resp3

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Decode reads from the provided bufio.Reader and interprets the next RESP3 data type,
// returning the parsed value as an interface{}. RESP3 protocol supports various data types
// like Simple Strings, Errors, Integers, Floats, Bulk Strings, Arrays, Booleans, Maps, and Nulls.
// This function is capable of decoding these types based on the initial byte that indicates
// the data type, followed by the data itself.
//
// Parameters:
//   - reader *bufio.Reader: A pointer to a bufio.Reader from which the data will be read. It is
//     expected that the reader is already initialized and points to a source of RESP3 formatted data.
//
// Returns:
//
//   - interface{}: The decoded data from the reader. The actual type of the returned value can be
//     one of several Go types depending on the RESP3 data type encountered. This could be a string
//     for Simple Strings and Bulk Strings, error for RESP3 Errors, int64 for Integers, float64 for
//     Floats, []interface{} for Arrays, map[string]interface{} for Maps, bool for Booleans, or nil
//     for Nulls.
//
//     Sample input string: "*5\r\n$4\r\nMSET\r\n$4\r\nkey1\r\n$16\r\nvalue1 dash dash\r\n$4\r\nkey2\r\n$6\r\nvalue2\r\n"
func Decode(reader *bufio.Reader) (interface{}, error) {
	dataType, err := reader.ReadByte()

	if err != nil {
		if err == io.EOF {
			return nil, io.EOF // Gracefully handle EOF by returning it
		}
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

		return errors.New(string(line)), nil

	case ':': // Integer
		line, _, err := reader.ReadLine()

		if err != nil {
			return nil, err
		}

		xint, castErr := strconv.Atoi(string(line))
		if castErr != nil {
			return xint, castErr
		}
		return int64(xint), nil

	case ',': // Float
		line, _, err := reader.ReadLine()

		if err != nil {
			return nil, err
		}

		xfloat, castErr := strconv.ParseFloat(string(line), 64)
		if castErr != nil {
			return xfloat, castErr
		}
		return float64(xfloat), nil

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
			element, err := Decode(reader)

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
			key, _ := Decode(reader)
			value, _ := Decode(reader)
			if key != nil {
				if key, ok := key.(string); ok {
					resultMap[key] = value
				}
			}
		}
		return resultMap, nil

	case '_':
		reader.Discard(2) // Discard the trailing \r\n
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported data type: %v", dataType)
	}
}
