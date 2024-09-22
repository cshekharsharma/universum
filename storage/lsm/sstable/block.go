package sstable

type Block struct {
	StartKey   string
	EndKey     string
	Data       map[string]interface{}
	Size       int64
	BlockIndex int64
}

func NewBlock() *Block {
	return &Block{
		Data: make(map[string]interface{}),
		Size: 0,
	}
}
