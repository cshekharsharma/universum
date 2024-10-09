package sstable

// sstIndexEntry is the structure of the object that is held in the SSTable index.
// key names have been intentionally kept short to reduce the size of the index, when
// it is stored on the disk in serialised format.
type sstIndexEntry struct {
	f string // first key
	l string // last key
	o int64  // block offset & length
}

func (si *sstIndexEntry) GetFirstKey() string {
	return si.f
}

func (si *sstIndexEntry) GetLastKey() string {
	return si.l
}

func (si *sstIndexEntry) GetOffset() int64 {
	return si.o
}

func (si *sstIndexEntry) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"f": si.f,
		"l": si.l,
		"o": si.o,
	}
}
