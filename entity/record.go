package entity

import (
	"time"
)

const (
	RecordTypeScalar = "scalar"
)

type Record interface {
	GetFamily() string
	IsExpired() bool
	ToMap() map[string]interface{}
	FromMap(m map[string]interface{}) (string, Record)
}

type ScalarRecord struct {
	Value  interface{}
	LAT    int64
	Expiry int64
}

func (sr *ScalarRecord) GetFamily() string {
	return RecordTypeScalar
}

func (sr *ScalarRecord) IsExpired() bool {
	if sr.Expiry == 0 {
		return false
	}

	return time.Now().Unix() > sr.Expiry
}

func (sr *ScalarRecord) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Value":  sr.Value,
		"LAT":    sr.LAT,
		"Expiry": sr.Expiry,
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

	return key, sr
}

type RecordKV struct {
	Key    string
	Record Record
}
