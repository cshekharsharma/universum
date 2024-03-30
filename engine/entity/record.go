package entity

type Record struct {
	TypeEncoding   uint8
	LastAccessedAt uint32
	Value          interface{}
}
