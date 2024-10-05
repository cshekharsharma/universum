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
}

type ScalarRecord struct {
	Value  interface{}
	Type   uint8
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
		"Type":   sr.Type,
		"LAT":    sr.LAT,
		"Expiry": sr.Expiry,
	}
}

type RecordKV struct {
	Key    string
	Record Record
}
