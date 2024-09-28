package lsm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
	"universum/config"
	"universum/internal/logger"
	"universum/resp3"
)

const (
	maxBufferSize    = 64 * 1024 * 1024 // 64MB buffer size
	maxFlushInterval = 5 * time.Second  // Maximum time to wait before flushing the buffer
	recoveryInterval = 2 * time.Second  // Time to wait before restarting the flusher
)

type WriteAheadLogger struct {
	file      *os.File
	buffer    *bytes.Buffer
	mu        sync.Mutex
	flusherCh chan struct{}
	ticker    *time.Ticker
}

// NewWAL initializes a new WAL instance.
func NewWAL(filedir string) (*WriteAheadLogger, error) {
	filePath := filepath.Clean(filedir + "/" + config.DefaultWALFileName)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return nil, fmt.Errorf("failed to open WAL file: %v", err)
	}

	walBufferSize := math.Min(float64(config.Store.Storage.LSM.WriteAheadLogBufferSize), maxBufferSize)
	flushInterval := time.Duration(math.Min(float64(config.Store.Storage.LSM.WriteAheadLogFrequency), float64(maxFlushInterval)))

	wal := &WriteAheadLogger{
		file:      file,
		buffer:    bytes.NewBuffer(make([]byte, 0, int(walBufferSize))),
		flusherCh: make(chan struct{}, 1),
		ticker:    time.NewTicker(flushInterval * time.Second),
	}

	go wal.startFlusher()

	return wal, nil
}

// AddToWALBuffer adds the key-value pair to the buffer.
func (wal *WriteAheadLogger) AddToWALBuffer(key string, value interface{}, ttl int64) error {
	wal.mu.Lock()
	defer wal.mu.Unlock()

	if err := wal.encodeLogEntry(key, value, ttl); err != nil {
		return fmt.Errorf("failed to encode log entry: %v", err)
	}

	if wal.buffer.Len() >= maxBufferSize {
		select {
		case wal.flusherCh <- struct{}{}:
		default:
		}
	}

	return nil
}

// encodeLogEntry encodes the key, value, and ttl in binary format.
func (wal *WriteAheadLogger) encodeLogEntry(key string, value interface{}, ttl int64) error {
	keyLen := int64(len(key))

	encodedValue, err := resp3.Encode(value)
	if err != nil {
		logger.Get().Error("AddToWALBuffer:: failed to resp-encode value for key %s, err=%v", key, err)
		return nil
	}

	valueBytes := []byte(encodedValue)
	valueLen := int64(len(valueBytes))

	if err := binary.Write(wal.buffer, binary.BigEndian, keyLen); err != nil {
		return err
	}
	if _, err := wal.buffer.WriteString(key); err != nil {
		return err
	}
	if err := binary.Write(wal.buffer, binary.BigEndian, valueLen); err != nil {
		return err
	}
	if _, err := wal.buffer.Write(valueBytes); err != nil {
		return err
	}
	if err := binary.Write(wal.buffer, binary.BigEndian, ttl); err != nil {
		return err
	}

	return nil
}

// startFlusher starts the WAL flusher that runs on a hybrid system (size and time-based flush).
func (wal *WriteAheadLogger) startFlusher() {
	for {
		select {
		case <-wal.ticker.C:
			wal.flush()

		case <-wal.flusherCh:
			wal.flush()
		}
	}
}

// flush writes the buffered data to the WAL file.
func (wal *WriteAheadLogger) flush() {
	defer func() {
		if r := recover(); r != nil {
			logger.Get().Warn("WAL flusher panicked, restarting...: %v", r)
			time.Sleep(recoveryInterval)
			go wal.startFlusher()
		}
	}()

	wal.mu.Lock()
	defer wal.mu.Unlock()

	if wal.buffer.Len() > 0 {
		_, err := wal.file.Write(wal.buffer.Bytes())
		if err != nil {
			fmt.Printf("failed to flush WAL buffer: %v\n", err)
			return
		}

		wal.buffer.Reset()

		err = wal.file.Sync()
		if err != nil {
			fmt.Printf("failed to sync WAL file: %v\n", err)
		}
	}
}

// Close closes the WAL file and stops the ticker.
func (wal *WriteAheadLogger) Close() {
	wal.ticker.Stop()
	wal.file.Close()
}
