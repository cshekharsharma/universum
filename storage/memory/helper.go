package memory

import (
	"universum/entity"
	"universum/utils"
)

func getRecordFromSerializedMap(recordMap map[string]interface{}) (string, entity.Record) {
	record := new(entity.ScalarRecord)
	var key string

	if k, ok := recordMap["Key"]; ok {
		key, _ = k.(string)
	}

	if val, ok := recordMap["Value"]; ok {
		record.Value = val
		record.Type = utils.GetTypeEncoding(val)
	}

	if lat, ok := recordMap["LAT"]; ok {
		record.LAT, _ = lat.(int64)
	}

	if expiry, ok := recordMap["Expiry"]; ok {
		record.Expiry, _ = expiry.(int64)
	}

	return key, record
}
