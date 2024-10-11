package wal

import "time"

const (
	maxBufferSize     = 64 * 1024 * 1024 // 64MB buffer size
	maxFlushInterval  = 30 * time.Second // Maximum time to wait before flushing the buffer
	recoveryInterval  = 2 * time.Second  // Time to wait before restarting the flusher
	maxWALFlushRetry  = 3                // Maximum number of retries for WAL flush
	fileSyncThreshold = 3                // Number of flushes before syncing the file
)
