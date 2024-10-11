package wal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/resp3"
	"universum/storage/lsm/memtable"
	"universum/utils"
)

type WALRecord struct {
	Key    string
	Value  interface{}
	Expiry int64
	State  uint8
}

type WALWriter struct {
	fileptr       *os.File
	buffer        *bytes.Buffer
	maxBufferSize int64
	mutex         sync.Mutex
	flusherCh     chan struct{}
	ticker        *time.Ticker
	isFlushing    bool
	syncCounter   int64
	syncThreshold int64
	walSize       int64
}

// NewWAL initializes a new WAL instance.
func NewWriter(filedir string) (*WALWriter, error) {
	filePath := filepath.Clean(filepath.Join(filedir, config.DefaultWALFileName))
	fileptr, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return nil, fmt.Errorf("failed to open WAL file: %v", err)
	}

	writer := &WALWriter{
		fileptr:       fileptr,
		isFlushing:    false,
		syncCounter:   0,
		syncThreshold: fileSyncThreshold,
		walSize:       0,
	}

	if config.Store.Storage.LSM.WriteAheadLogAsyncFlush {
		cnf := config.Store.Storage.LSM
		writer.maxBufferSize = int64(math.Min(float64(cnf.WriteAheadLogBufferSize), maxBufferSize))
		flushInterval := time.Duration(math.Min(float64(cnf.WriteAheadLogFrequency), float64(maxFlushInterval)))

		writer.buffer = bytes.NewBuffer(make([]byte, 0, writer.maxBufferSize))
		writer.ticker = time.NewTicker(flushInterval * time.Second)
		writer.flusherCh = make(chan struct{}, 1)

		go writer.startFlusher()
	}

	return writer, nil
}

// AddToWALBuffer adds the key-value pair to the buffer.
func (ww *WALWriter) AddToWALBuffer(key string, value interface{}, ttl int64, state uint8) error {
	ww.mutex.Lock()
	defer ww.mutex.Unlock()

	select {
	case <-memtable.WALRotaterChan:
		ww.RotateWALFile()

	default:
	}

	encodedCommand, err := ww.getEncodedEntries(key, value, ttl, state)
	if err != nil {
		return fmt.Errorf("AddToWALBuffer:: WAL append failed: %v", err)
	}

	commandBytes := []byte(encodedCommand)
	commandLen := int64(len(commandBytes))

	// when WriteAheadLogAsyncFlush is false, then writer will write to file immediately
	// without calling fsync. Otherwise it'll just append to the buffer and writer to file
	// through an async worker, and commit fsync immediately.
	if !config.Store.Storage.LSM.WriteAheadLogAsyncFlush {
		buf := bytes.NewBuffer(make([]byte, 0, commandLen+entity.Int64SizeInBytes))
		if err := binary.Write(buf, binary.BigEndian, commandLen); err != nil {
			return err
		}

		if _, err := buf.Write(commandBytes); err != nil {
			return err
		}

		_, err := ww.fileptr.Write(buf.Bytes())
		if err != nil {
			return fmt.Errorf("AddToWALBuffer: WAL append failed: %v", err)
		}

		return nil
	} else {
		if err := binary.Write(ww.buffer, binary.BigEndian, commandLen); err != nil {
			return err
		}

		if _, err := ww.buffer.Write(commandBytes); err != nil {
			return err
		}
	}

	if int64(ww.buffer.Len()) >= ww.maxBufferSize && !ww.isFlushing {
		ww.isFlushing = true

		select {
		case ww.flusherCh <- struct{}{}:
		default:
		}
	}

	return nil
}

// logEncodedEntries encodes the key, value, and other params into the buffer.
func (ww *WALWriter) getEncodedEntries(key string, value interface{}, ttl int64, state uint8) (string, error) {
	var command = make(map[string]interface{})

	expiry := utils.GetCurrentEPochTime() + ttl
	if ttl == 0 {
		expiry = config.InfiniteExpiryTime
	}

	command = map[string]interface{}{
		"Key":    key,
		"Value":  value,
		"Expiry": expiry,
		"State":  state,
	}

	encodedCommand, err := resp3.Encode(command)
	if err != nil {
		return "", fmt.Errorf("failed to resp-encode command for key %s, err=%v", key, err)
	}

	return encodedCommand, nil
}

// startFlusher starts the WAL flusher that runs on a hybrid system (size and time-based flush).
func (ww *WALWriter) startFlusher() {
	for {
		select {
		case <-ww.ticker.C:
			ww.flush()

		case <-ww.flusherCh:
			ww.flush()
		}
	}
}

// flush writes the buffered data to the WAL file.
func (ww *WALWriter) flush() {
	defer func() {
		if r := recover(); r != nil {
			ww.isFlushing = false
			logger.Get().Warn("WAL flusher panicked, restarting...: %v", r)
			time.Sleep(recoveryInterval)
			go ww.startFlusher()
		}
	}()

	ww.mutex.Lock()
	defer ww.mutex.Unlock()

	bufLength := ww.buffer.Len()
	if ww.buffer.Len() > 0 {
		isFlushed := false
		retryCount := 0

		for retryCount < maxWALFlushRetry {
			retryCount++
			err := ww.attemptFlush(retryCount)
			if err == nil {
				isFlushed = true
				break
			}

			// Exponential backoff to avoid flooding the file writes on failureF
			time.Sleep((1 << retryCount) * 10 * time.Millisecond)
			continue
		}

		if !isFlushed {
			logger.Get().Error("Failed to flush WAL buffer into file after %d retries", retryCount)
			return
		}

		ww.buffer.Next(bufLength)
		ww.syncCounter++
		ww.isFlushing = false

		if ww.syncCounter >= ww.syncThreshold {
			err := ww.fileptr.Sync()
			if err != nil {
				logger.Get().Error("failed to sync WAL file: %v\n", err)
			}
			ww.syncCounter = 0
		}
	}
}

// attemptFlush writes the buffer to the file.
func (ww *WALWriter) attemptFlush(retryCount int) error {
	_, err := ww.fileptr.Write(ww.buffer.Bytes())
	if err != nil {
		logger.Get().Error("[#%d] failed to flush WAL buffer: %v", retryCount, err)
		return err
	}
	return nil
}

// RotateWALFile truncates the WAL file to zero bytes.
func (ww *WALWriter) RotateWALFile() error {
	err := ww.fileptr.Truncate(0)
	if err != nil {
		return err
	}

	_, err = ww.fileptr.Seek(0, io.SeekStart)
	if err != nil {
		return nil
	}

	return nil
}

// Close closes the WAL file and stops the ticker.
func (ww *WALWriter) Close() {
	if config.Store.Storage.LSM.WriteAheadLogAsyncFlush {
		if ww.ticker != nil {
			ww.ticker.Stop()
		}

		ww.flush() // final flush before closing
	}

	err := ww.fileptr.Close()
	if err != nil {
		logger.Get().Error("Failed to close WAL file: %v", err)
	}
}
