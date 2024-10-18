package entity

import (
	"time"
)

const (
	RecordTypeScalar = "scalar"

	RecordStateActive     = 0
	RecordStateTombstoned = 1
	RecordStateObsolete   = 2
)

type Record interface {
	GetFamily() string
	GetValue() interface{}
	GetExpiry() int64
	IsExpired() bool
	IsTombstoned() bool
	ToMap() map[string]interface{}
	FromMap(m map[string]interface{}) (string, Record)
}

type ScalarRecord struct {
	Value  interface{}
	LAT    int64
	Expiry int64
	State  uint8
}

func (sr *ScalarRecord) GetFamily() string {
	return RecordTypeScalar
}

func (sr *ScalarRecord) GetValue() interface{} {
	return sr.Value
}

func (sr *ScalarRecord) GetExpiry() int64 {
	return sr.Expiry
}

func (sr *ScalarRecord) IsExpired() bool {
	if sr.Expiry == 0 {
		return false
	}

	return time.Now().Unix() > sr.Expiry
}

func (sr *ScalarRecord) IsTombstoned() bool {
	return sr.State == RecordStateTombstoned
}

func (sr *ScalarRecord) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Value":  sr.Value,
		"LAT":    sr.LAT,
		"Expiry": sr.Expiry,
		"State":  sr.State,
	}
}

func (sr *ScalarRecord) FromMap(recordMap map[string]interface{}) (string, Record) {
	var key string

	if k, ok := recordMap["Key"]; ok {
		key, _ = k.(string)
	}

	if val, ok := recordMap["Value"]; ok {
		sr.Value = val
	}

	if lat, ok := recordMap["LAT"]; ok {
		sr.LAT, _ = lat.(int64)
	}

	if expiry, ok := recordMap["Expiry"]; ok {
		sr.Expiry, _ = expiry.(int64)
	}

	if state, ok := recordMap["State"]; ok {
		sr.State, _ = state.(uint8)
	}

	return key, sr
}

type RecordKV struct {
	Key    string
	Record Record
}
