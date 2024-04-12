package storage

import (
	"time"
)

const (
	RECORD_TYPE_SCALER = "scalar"
)

type Record interface {
	GetFamily() string
	IsExpired() bool
}

type ScalarRecord struct {
	Value  interface{}
	Type   uint8
	LAT    uint32
	Expiry int64
}

func (sr *ScalarRecord) GetFamily() string {
	return RECORD_TYPE_SCALER
}

func (sr *ScalarRecord) IsExpired() bool {
	if sr.Expiry == 0 {
		return false
	}

	return time.Now().Unix() > sr.Expiry
}
