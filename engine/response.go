package engine

import (
	"fmt"
	"universum/utils"
)

const (
	RESPONSECODE_RECORD_FOUND uint32 = 1000
	RESPONSECODE_UPDATED      uint32 = 1010

	RESPONSECODE_INVALID_CMD_INPUT uint32 = 3000
	RESPONSECODE_KEY_NOT_FOUND     uint32 = 3010
	RESPONSECODE_KEY_EXPIRED       uint32 = 3020
)

func EncodedResponse(response interface{}) string {
	encoded, err := utils.EncodeRESP3(response)

	if err != nil {
		newErr := fmt.Errorf("unexpected error occured while processing output: %v", err)
		encoded, _ = utils.EncodeRESP3(newErr)
	}

	return encoded
}
