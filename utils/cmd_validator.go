package utils

import (
	"fmt"
	"reflect"
	"universum/entity"
)

type ValidationRule struct {
	Name     string
	Datatype reflect.Kind
}

func ValidateArguments(cmd *entity.Command, validations []ValidationRule) (bool, []interface{}) {
	requiredCount := len(validations)
	argumentCount := len(cmd.Args)

	if argumentCount != requiredCount {
		return false, []interface{}{
			nil,
			entity.CRC_INVALID_CMD_INPUT,
			fmt.Sprintf("ERR: Incorrect number of arguments provided. Want=%d, Have=%d",
				requiredCount, argumentCount),
		}
	}

	for i := 0; i < argumentCount; i++ {
		argument := cmd.Args[i]

		// wildcard datatype where we dont care about the type
		if validations[i].Datatype == reflect.Interface {
			continue
		}

		if reflect.TypeOf(argument).Kind() != validations[i].Datatype {
			return false, []interface{}{
				nil,
				entity.CRC_INVALID_CMD_INPUT,
				fmt.Sprintf("ERR: %s has invalid type. %s expected", validations[i].Name, validations[i].Datatype),
			}
		}
	}

	return true, []interface{}{}
}
