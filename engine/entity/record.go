package entity

type Record struct {
	Value interface{}
	Type  uint8
	LAT   uint32
}

func NewRecord(value interface{}, typeEncoding uint8, lat uint32) *Record {
	record := new(Record)

	record.Value = value
	record.Type = typeEncoding
	record.LAT = lat

	return record
}
