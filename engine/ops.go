// lists and implements all operations supported by the database
package engine

import (
	"fmt"
	"universum/engine/entity"
)

var storage Storage

func Shutdown() {
	// do all the shut down operations, such as fsyncing AOF
	// and freeing up occupied resources and memory.
}

func Startup() {
	storage := new(Storage)
	storage.Initialize()
}

func executeGET(command *entity.Command) string {
	var output []interface{}

	if len(command.Args) < 1 {
		output = []interface{}{
			nil,
			RESPONSECODE_INVALID_CMD_INPUT,
			"ERR: too few arguments provided. Want=1, Have=0",
		}

		return EncodedResponse(output)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			nil,
			RESPONSECODE_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return EncodedResponse(output)
	}

	record, code := storage.Get(key)

	return EncodedResponse([]interface{}{
		record,
		code,
		"",
	})
}

func executeSET(command *entity.Command) string {
	var output []interface{}

	argLength := len(command.Args)
	if argLength != 3 {
		output = []interface{}{
			false,
			RESPONSECODE_INVALID_CMD_INPUT,
			fmt.Sprintf("ERR: Incorrect number of args provided. Want=3, Have=%d", argLength),
		}

		return EncodedResponse(output)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			false,
			RESPONSECODE_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return EncodedResponse(output)
	}

	ttl, ok := command.Args[2].(int)
	if !ok {
		output = []interface{}{
			false,
			RESPONSECODE_INVALID_CMD_INPUT,
			"ERR: TTL has invalid type, uint64 expected",
		}

		return EncodedResponse(output)
	}

	success, code := storage.Set(key, command.Args[1], uint64(ttl))

	return EncodedResponse([]interface{}{
		success,
		code,
		"",
	})
}
