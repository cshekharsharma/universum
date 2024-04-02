package entity

type Record struct {
	Value          interface{}
	TypeEncoding   uint8
	LastAccessedAt uint32
}

func NewRecord(value interface{}, typeEncoding uint8, lat uint32) *Record {
	record := new(Record)

	record.Value = value
	record.TypeEncoding = typeEncoding
	record.LastAccessedAt = lat

	return record
}
