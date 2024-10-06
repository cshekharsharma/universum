package memtable

// FlusherChan receives a message from Memtable if that memtable is full
// and needs to be flushed to SSTable on disk by the LSM engine.
// Listener to this flusher channel is implemented in LSMTree
var FlusherChan chan MemTable

// WALRotaterChan receives a message from Memtable when it's ready to flush
// the memtable to disk. This message indicates that the WAL can be rotated
var WALRotaterChan chan int64
