package resp3

import "fmt"

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
	encoded, err := Encode(response)

	if err != nil {
		newErr := fmt.Errorf("unexpected error occured while processing output: %v", err)
		encoded, _ = Encode(newErr)
	}

	return encoded
}
