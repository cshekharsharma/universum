package memtable

// FlusherChan receives a message from Memtable if that memtable is full
// and needs to be flushed to SSTable on disk by the LSM engine.
// Listener to this flusher channel is implemented in LSMTree
var FlusherChan chan MemTable
