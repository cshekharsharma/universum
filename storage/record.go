package storage

import (
	"universum/utils"
)

const (
	RECORD_TYPE_SCALER = "scalar"
)

type Record interface {
	GetFamily() string
	IsExpired() bool
}

type RecordResponse struct {
	Record Record
	Code   uint32
}

type ScalarRecord struct {
	Value  interface{}
	Type   uint8
	LAT    int64
	Expiry int64
}

func (sr *ScalarRecord) GetFamily() string {
	return RECORD_TYPE_SCALER
}

func (sr *ScalarRecord) IsExpired() bool {
	if sr.Expiry == 0 {
		return false
	}

	return utils.GetCurrentEPochTime() > sr.Expiry
}
