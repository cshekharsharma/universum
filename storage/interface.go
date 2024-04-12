package storage

type DataStore interface {
	Initialize()
	Exists(key string) (bool, uint32)
	Get(key string) (Record, uint32)
	Set(key string, value interface{}, ttl int64) (bool, uint32)
	Delete(key string) (bool, uint32)
	IncrDecrInteger(key string, offset int64, isIncr bool) (int64, uint32)
	Append(key string, value string) (int64, uint32)
}
