package engine

import (
	"bufio"
	"fmt"
	"strconv"
)

// decodeRESP3 decodes a RESP3 encoded message from a bufio.Reader.
// It returns the decoded message as an interface{}, which can be type asserted as needed.
//
// # Sample code for using RESP3 parser
//
//	input := []byte("*5\r\n$4\r\nMSET\r\n$4\r\nkey1\r\n$16\r\nvalue1 dash dash\r\n$4\r\nkey2\r\n$6\r\nvalue2\r\n")
//
//	reader := bufio.NewReader(bytes.NewReader(input))
//	decoded, err := decodeRESP3(reader)
//	if err != nil {
//		fmt.Println("Error decoding:", err)
//		return
//	}
//	fmt.Printf("Decoded: %#v\n", decoded)
func decodeRESP3(reader *bufio.Reader) (interface{}, error) {
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

		// Read the trailing CRLF
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
			element, err := decodeRESP3(reader)

			if err != nil {
				return nil, err
			}

			array[i] = element
		}

		return array, nil
	default:
		return nil, fmt.Errorf("unsupported data type: %v", dataType)
	}
}
