package utils

import (
	"fmt"
	"universum/resp3"
)

func EncodedRESP3Response(response interface{}) string {
	encoded, err := resp3.Encode(response)

	if err != nil {
		newErr := fmt.Errorf("unexpected error occured while processing output: %v", err)
		encoded, _ = resp3.Encode(newErr)
	}

	return encoded
}
