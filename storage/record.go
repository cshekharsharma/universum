package storage

import (
	"universum/utils"
)

const (
	RecordTypeScalar = "scalar"
)

type Record interface {
	GetFamily() string
	IsExpired() bool
}

type RecordResponse struct {
	Value interface{}
	Code  uint32
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

	return utils.GetCurrentEPochTime() > sr.Expiry
}
