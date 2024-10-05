package memory

import (
	"reflect"
	"testing"
	"universum/entity"
	"universum/utils"
)

func TestGetRecordFromSerializedMap(t *testing.T) {
	tests := []struct {
		name       string
		recordMap  map[string]interface{}
		wantKey    string
		wantRecord entity.Record
	}{
		{
			name: "ValidRecordMap",
			recordMap: map[string]interface{}{
				"Key":    "testKey",
				"Value":  "testValue",
				"LAT":    int64(1234567890),
				"Expiry": int64(9876543210),
			},
			wantKey: "testKey",
			wantRecord: &entity.ScalarRecord{
				Value:  "testValue",
				Type:   utils.GetTypeEncoding("testValue"),
				LAT:    int64(1234567890),
				Expiry: int64(9876543210),
			},
		},
		{
			name: "MissingOptionalFields",
			recordMap: map[string]interface{}{
				"Key":   "testKey",
				"Value": "testValue",
			},
			wantKey: "testKey",
			wantRecord: &entity.ScalarRecord{
				Value: "testValue",
				Type:  utils.GetTypeEncoding("testValue"),
			},
		},
		{
			name: "EmptyRecordMap",
			recordMap: map[string]interface{}{
				"Key": "",
			},
			wantKey:    "",
			wantRecord: &entity.ScalarRecord{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotRecord := getRecordFromSerializedMap(tt.recordMap)
			if gotKey != tt.wantKey {
				t.Errorf("getRecordFromSerializedMap() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
			if !reflect.DeepEqual(gotRecord, tt.wantRecord) {
				t.Errorf("getRecordFromSerializedMap() gotRecord = %v, want %v", gotRecord, tt.wantRecord)
			}
		})
	}
}
