// lists and implements all operations supported by the database
package engine

import (
	"reflect"
	"universum/config"
	"universum/entity"
	"universum/resp3"
	"universum/utils"
)

type Command struct {
	Name string
	Args []interface{}
}

func executePING(command *entity.Command) string {
	rules := []utils.ValidationRule{}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	return resp3.EncodedRESP3Response([]interface{}{"OK", entity.CRC_PING_SUCCESS, ""})
}

func executeEXISTS(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	exists, code := datastore.Exists(key)

	return resp3.EncodedRESP3Response([]interface{}{exists, code, ""})
}

func executeGET(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	record, code := datastore.Get(key)

	var recordAsMap interface{} = record
	if record != nil {
		recordAsMap = record.ToMap()
	}
	return resp3.EncodedRESP3Response([]interface{}{recordAsMap, code, ""})
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

	success, code := datastore.Set(key, value, ttl)
	return resp3.EncodedRESP3Response([]interface{}{success, code, ""})
}

func executeDELETE(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	deleted, code := datastore.Delete(key)

	return resp3.EncodedRESP3Response([]interface{}{deleted, code, ""})
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

	updatedValue, code := datastore.IncrDecrInteger(key, offset, isIncr)
	return resp3.EncodedRESP3Response([]interface{}{updatedValue, code, ""})
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

	length, code := datastore.Append(key, value)
	return resp3.EncodedRESP3Response([]interface{}{length, code, ""})
}

func executeMGET(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "keys", Datatype: reflect.Slice},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	keyIntrSlice, _ := command.Args[0].([]interface{})
	keyStringSlice := make([]string, 0, len(keyIntrSlice))

	for kk := range keyIntrSlice {
		val, isOk := keyIntrSlice[kk].(string)

		if !isOk {
			return resp3.EncodedRESP3Response([]interface{}{
				nil, entity.CRC_INVALID_CMD_INPUT,
				"first argument should be a list of string, one or more invalid values provided"})
		}

		keyStringSlice = append(keyStringSlice, val)
	}

	records, code := datastore.MGet(keyStringSlice)
	return resp3.EncodedRESP3Response([]interface{}{records, code, ""})
}

func executeMSET(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "KvMap", Datatype: reflect.Map},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	kvMap, ok := command.Args[0].(map[string]interface{})

	if !ok {
		return resp3.EncodedRESP3Response([]interface{}{
			nil, entity.CRC_INVALID_CMD_INPUT,
			"first argument should be a dict of string to anything, one or more invalid values provided"})
	}

	setStatuses, code := datastore.MSet(kvMap)
	return resp3.EncodedRESP3Response([]interface{}{setStatuses, code, ""})
}

func executeMDELETE(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "keys", Datatype: reflect.Slice},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	keyIntrSlice, _ := command.Args[0].([]interface{})
	keyStringSlice := make([]string, 0, len(keyIntrSlice))

	for kk := range keyIntrSlice {
		val, isOk := keyIntrSlice[kk].(string)

		if !isOk {
			return resp3.EncodedRESP3Response([]interface{}{
				nil, entity.CRC_INVALID_CMD_INPUT,
				"first argument should be a list of string, one or more invalid values provided"})
		}

		keyStringSlice = append(keyStringSlice, val)
	}

	deleteStatuses, code := datastore.MDelete(keyStringSlice)
	return resp3.EncodedRESP3Response([]interface{}{deleteStatuses, code, ""})
}

func executeTTL(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	ttl, code := datastore.TTL(key)

	return resp3.EncodedRESP3Response([]interface{}{ttl, code, ""})
}

func executeEXPIRE(command *entity.Command) string {
	rules := []utils.ValidationRule{
		{Name: "key", Datatype: reflect.String},
		{Name: "ttl", Datatype: reflect.Int64},
	}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	key, _ := command.Args[0].(string)
	ttl, _ := command.Args[1].(int64)

	success, code := datastore.Expire(key, ttl)
	return resp3.EncodedRESP3Response([]interface{}{success, code, ""})

}
func executeSNAPSHOT(command *entity.Command) string {
	rules := []utils.ValidationRule{}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	err := StartDatabaseSnapshot(getDataStore(config.Store.Storage.StorageEngine))
	if err != nil {
		return resp3.EncodedRESP3Response([]interface{}{false, entity.CRC_SNAPSHOT_FAILED, err.Error()})
	}

	return resp3.EncodedRESP3Response([]interface{}{true, entity.CRC_SNAPSHOT_STARTED, ""})
}

func executeINFO(command *entity.Command) string {
	rules := []utils.ValidationRule{}

	if isValid, validityRes := utils.ValidateArguments(command, rules); !isValid {
		return resp3.EncodedRESP3Response(validityRes)
	}

	infoResponse := GetDatabaseInfoStatistics().ToString()
	return resp3.EncodedRESP3Response([]interface{}{infoResponse, entity.CRC_INFO_CONTENT_OK, ""})
}

func executeHELP(command *entity.Command) string {
	helpcontent := ""
	argLength := len(command.Args)

	if argLength == 0 {
		helpcontent = getGenericHelpContent()
	}

	if argLength == 1 {
		commandName, ok := command.Args[0].(string)
		if !ok {
			return resp3.EncodedRESP3Response([]interface{}{
				nil, entity.CRC_INVALID_CMD_INPUT,
				"first argument should be a valid db command"})
		}

		helpcontent = getCommandHelpContent(commandName)
	}

	return resp3.EncodedRESP3Response([]interface{}{helpcontent, entity.CRC_HELP_CONTENT_OK, ""})
}
