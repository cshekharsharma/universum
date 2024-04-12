package engine

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"universum/internal/logger"
)

var translogBufferInstance *translogBuffer
var translogInitMutex sync.Mutex

type translogBuffer struct {
	buffer bytes.Buffer
	mutex  sync.Mutex
}

func newRecordTranslogBuffer() *translogBuffer {
	if translogBufferInstance == nil {
		translogInitMutex.Lock()

		if translogBufferInstance == nil {
			translogBufferInstance = new(translogBuffer)
		}

		translogInitMutex.Unlock()
	}

	return translogBufferInstance
}

func (tb *translogBuffer) addToTranslogBuffer(message string) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	_, err := tb.buffer.WriteString(message)

	if err != nil {
		logger.Get().Warn("Failed to write message into translog buffer: %v", err.Error())
		return false
	}

	return true
}

func (tb *translogBuffer) flush(targetFile string) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.buffer.Len() == 0 {
		return
	}

	tb.flushBufferToFile(&tb.buffer, targetFile)
	tb.buffer.Reset()
}

func (tb *translogBuffer) flushBufferToFile(buffer *bytes.Buffer, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return fmt.Errorf("FileError: %v", err)
	}
	defer file.Close()

	if _, err := buffer.WriteTo(file); err != nil {
		return fmt.Errorf("FileError: %v", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("FileError: %v", err)
	}

	return nil
}
