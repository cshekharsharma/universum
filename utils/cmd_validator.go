package utils

import (
	"fmt"
	"universum/consts"
	"universum/engine/entity"
)

func ValidateArgumentCount(cmd *entity.Command, requiredCount int) (bool, []interface{}) {
	hasError := false
	output := []interface{}{}

	if len(cmd.Args) != requiredCount {
		hasError = true
		return true, []interface{}{
			nil,
			consts.CRC_INVALID_CMD_INPUT,
			fmt.Sprintf("ERR: Incorrect number of arguments provided. Want=%d, Have=%d",
				requiredCount, len(cmd.Args)),
		}
	}

	return hasError, output
}
