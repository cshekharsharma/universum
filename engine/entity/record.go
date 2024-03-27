package entity

type Record struct {
	TypeEncoding   uint8
	LastAccessedAt int32
	Value          interface{}
}
