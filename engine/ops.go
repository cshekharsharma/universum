// lists and implements all operations supported by the database
package engine

import (
	"universum/consts"
	"universum/engine/entity"
	"universum/utils"
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

	if hasError, validityRes := utils.ValidateArgumentCount(command, 1); hasError {
		return utils.EncodedResponse(validityRes)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			nil,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return utils.EncodedResponse(output)
	}

	record, code := storage.Get(key)

	return utils.EncodedResponse([]interface{}{
		record,
		code,
		"",
	})
}

func executeSET(command *entity.Command) string {
	var output []interface{}

	if hasError, validityRes := utils.ValidateArgumentCount(command, 3); hasError {
		return utils.EncodedResponse(validityRes)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			false,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return utils.EncodedResponse(output)
	}

	ttl, ok := command.Args[2].(int64)
	if !ok {
		output = []interface{}{
			false,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: TTL has invalid type, int64 expected",
		}

		return utils.EncodedResponse(output)
	}

	success, code := storage.Set(key, command.Args[1], uint32(ttl))

	return utils.EncodedResponse([]interface{}{
		success,
		code,
		"",
	})
}

func executeEXISTS(command *entity.Command) string {
	var output []interface{}

	if hasError, validityRes := utils.ValidateArgumentCount(command, 1); hasError {
		return utils.EncodedResponse(validityRes)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			nil,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return utils.EncodedResponse(output)
	}

	exists, code := storage.Exists(key)

	return utils.EncodedResponse([]interface{}{
		exists,
		code,
		"",
	})
}

func executeDELETE(command *entity.Command) string {
	var output []interface{}

	if hasError, validityRes := utils.ValidateArgumentCount(command, 1); hasError {
		return utils.EncodedResponse(validityRes)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			nil,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return utils.EncodedResponse(output)
	}

	deleted, code := storage.Delete(key)

	return utils.EncodedResponse([]interface{}{
		deleted,
		code,
		"",
	})
}

func executeINCRDECR(command *entity.Command, isIncr bool) string {
	var output []interface{}

	if hasError, validityRes := utils.ValidateArgumentCount(command, 2); hasError {
		return utils.EncodedResponse(validityRes)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			false,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return utils.EncodedResponse(output)
	}

	offset, ok := command.Args[1].(int64)
	if !ok {
		output = []interface{}{
			false,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: Offset has invalid type, int64 expected",
		}

		return utils.EncodedResponse(output)
	}

	updatedValue, code := storage.IncrDecrInteger(key, int64(offset), isIncr)

	return utils.EncodedResponse([]interface{}{
		updatedValue,
		code,
		"",
	})
}
