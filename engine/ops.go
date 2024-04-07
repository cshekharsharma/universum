// lists and implements all operations supported by the database
package engine

import (
	"log"
	"os"
	"universum/config"
	"universum/consts"
	"universum/engine/entity"
	"universum/resp3"
	"universum/storage"
	"universum/utils"
)

var memstore *storage.MemoryStore

func GetMemstore() *storage.MemoryStore {
	return memstore
}

func Startup() {
	// Initiaise the in-memory store
	memstore = storage.CreateNewMemoryStore()
	memstore.Initialize()

	// Replay all commands from translog into the database
	keyCount, err := ReplayTranslog(config.GetForceAOFReplayOnError())
	if err != nil {
		log.Printf("[Translog Replay [failed]: %v\n [KEYOFFSET: %d]", err, keyCount+1)
		Shutdown()
	} else {
		log.Printf("[Translog Replay [success]: Total %d keys successfully replayed into the database.\n", keyCount)
	}

	// Trigger a expiry background job to periodically
	// delete expired keys from the database
	triggerPeriodicExpiryJob()
}

func executeGET(command *entity.Command) string {
	var output []interface{}

	if hasError, validityRes := utils.ValidateArgumentCount(command, 1); hasError {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			nil,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return resp3.EncodedRESP3Response(output)
	}

	record, code := memstore.Get(key)

	return resp3.EncodedRESP3Response([]interface{}{
		record,
		code,
		"",
	})
}

func executeSET(command *entity.Command) string {
	var output []interface{}

	if hasError, validityRes := utils.ValidateArgumentCount(command, 3); hasError {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			false,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return resp3.EncodedRESP3Response(output)
	}

	ttl, ok := command.Args[2].(int64)
	if !ok {
		output = []interface{}{
			false,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: TTL has invalid type, int64 expected",
		}

		return resp3.EncodedRESP3Response(output)
	}

	success, code := memstore.Set(key, command.Args[1], ttl)

	if success && code == consts.CRC_RECORD_UPDATED {
		NewTranslogBuffer().AddToBuffer(string(command.Raw))
	}

	return resp3.EncodedRESP3Response([]interface{}{
		success,
		code,
		"",
	})
}

func executeEXISTS(command *entity.Command) string {
	var output []interface{}

	if hasError, validityRes := utils.ValidateArgumentCount(command, 1); hasError {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			nil,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return resp3.EncodedRESP3Response(output)
	}

	exists, code := memstore.Exists(key)

	return resp3.EncodedRESP3Response([]interface{}{
		exists,
		code,
		"",
	})
}

func executeDELETE(command *entity.Command) string {
	var output []interface{}

	if hasError, validityRes := utils.ValidateArgumentCount(command, 1); hasError {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			nil,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return resp3.EncodedRESP3Response(output)
	}

	deleted, code := memstore.Delete(key)

	if deleted && code == consts.CRC_RECORD_DELETED {
		NewTranslogBuffer().AddToBuffer(string(command.Raw))
	}

	return resp3.EncodedRESP3Response([]interface{}{
		deleted,
		code,
		"",
	})
}

func executeINCRDECR(command *entity.Command, isIncr bool) string {
	var output []interface{}

	if hasError, validityRes := utils.ValidateArgumentCount(command, 2); hasError {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, ok := command.Args[0].(string)
	if !ok {
		output = []interface{}{
			false,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: key has invalid type. string expected",
		}

		return resp3.EncodedRESP3Response(output)
	}

	offset, ok := command.Args[1].(int64)
	if !ok {
		output = []interface{}{
			false,
			consts.CRC_INVALID_CMD_INPUT,
			"ERR: Offset has invalid type, int64 expected",
		}

		return resp3.EncodedRESP3Response(output)
	}

	updatedValue, code := memstore.IncrDecrInteger(key, int64(offset), isIncr)

	return resp3.EncodedRESP3Response([]interface{}{
		updatedValue,
		code,
		"",
	})
}

func Shutdown() {
	// do all the shut down operations, such as fsyncing AOF
	// and freeing up occupied resources and memory.
	NewTranslogBuffer().Flush()
	os.Exit(0)
}
