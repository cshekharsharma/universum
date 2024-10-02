package resp3

import (
	"bufio"
	"bytes"
	"fmt"
	"universum/entity"

	"github.com/cshekharsharma/resp-go/resp3"
)

// EncodedRESP3Response takes a generic response object, attempts to encode it into
// the RESP3 format using the Encode function, and returns the encoded string.
// If an error occurs during encoding, it captures the error, encodes the error message
// into RESP3 format, and returns the error message as a RESP3 encoded string.
//
// Parameters:
//   - response interface{}: The response to be encoded into RESP3 format. This can be
//     any of the supported types by the Encode function, including basic Go types, slices,
//     maps, custom structs, and even nil values.
//
// Returns:
//   - string: The RESP3 encoded string representation of the input `response`. If the
//     input is an error or encoding fails, the returned string will be an encoded error message.
func EncodedRESP3Response(response interface{}) string {
	encoded, err := resp3.Encode(response)

	if err != nil {
		newErr := fmt.Errorf("unexpected error occured while processing output: %v", err)
		encoded, _ = resp3.Encode(newErr)
	}

	return encoded
}

// Wrapper function for resp3.Encode
func Encode(value interface{}) (string, error) {
	if _, ok := value.(*entity.ScalarRecord); ok {
		return resp3.Encode(value.(*entity.ScalarRecord).ToMap())
	}

	return resp3.Encode(value)
}

// Wrapper function for resp3.Decode
func Decode(reader *bufio.Reader) (interface{}, error) {
	return resp3.Decode(reader)
}

func GetScalarRecordFromResp(raw string) (entity.Record, error) {
	decodedRecord, err := resp3.Decode(bufio.NewReader(bytes.NewReader([]byte(raw))))
	if err != nil {
		return nil, fmt.Errorf("record could not be decoded: %v", err)
	}

	if record, ok := decodedRecord.(map[string]interface{}); ok {
		return &entity.ScalarRecord{
			Value:  record["Value"],
			Type:   uint8(record["Type"].(int64)),
			LAT:    int64(record["LAT"].(int64)),
			Expiry: int64(record["Expiry"].(int64)),
		}, nil
	}

	return nil, fmt.Errorf("record is not in the correct format: %v", decodedRecord)
}
