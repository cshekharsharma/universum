// lists and implements all operations supported by the database
package engine

import (
	"os"
	"reflect"
	"universum/engine/entity"
	"universum/internal/logger"
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
	keyCount, err := PopulateRecordsFromSnapshot()

	if err != nil {
		logger.Get().Error("Translog replay failed, KeyOffset=%d, Err=%v", keyCount+1, err.Error())
		Shutdown()
	} else {
		logger.Get().Info("Translog replay done. Total %d keys replayed into DB", keyCount)
	}

	// Trigger periodic jobs
	triggerPeriodicExpiryJob()
	triggerPeriodicSnapshotJob()
}

func executeGET(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	record, code := memstore.Get(key)

	return resp3.EncodedRESP3Response([]interface{}{
		record,
		code,
		"",
	})
}

func executeSET(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
		{Name: "value", Datatype: reflect.Interface},
		{Name: "ttl", Datatype: reflect.Int64},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	value := command.Args[1]
	ttl, _ := command.Args[2].(int64)

	success, code := memstore.Set(key, value, ttl)

	return resp3.EncodedRESP3Response([]interface{}{
		success,
		code,
		"",
	})
}

func executeEXISTS(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	exists, code := memstore.Exists(key)

	return resp3.EncodedRESP3Response([]interface{}{
		exists,
		code,
		"",
	})
}

func executeDELETE(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	deleted, code := memstore.Delete(key)

	return resp3.EncodedRESP3Response([]interface{}{
		deleted,
		code,
		"",
	})
}

func executeINCRDECR(command *entity.Command, isIncr bool) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
		{Name: "offset", Datatype: reflect.Int64},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	offset, _ := command.Args[1].(int64)

	updatedValue, code := memstore.IncrDecrInteger(key, offset, isIncr)

	return resp3.EncodedRESP3Response([]interface{}{
		updatedValue,
		code,
		"",
	})
}

func executeAPPEND(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
		{Name: "value", Datatype: reflect.String},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	value, _ := command.Args[1].(string)

	length, code := memstore.Append(key, value)

	return resp3.EncodedRESP3Response([]interface{}{
		length,
		code,
		"",
	})
}

func Shutdown() {
	// do all the shut down operations, such as fsyncing AOF
	// and freeing up occupied resources and memory.
	StartInMemoryDBSnapshot(GetMemstore())
	os.Exit(0)
}
