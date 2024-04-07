package engine

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"universum/config"
	"universum/resp3"
)

var translogInstance *transLogBuffer
var translogMutex sync.Mutex

type transLogBuffer struct {
	buffer bytes.Buffer
	mutex  sync.Mutex
}

func NewTranslogBuffer() *transLogBuffer {
	if translogInstance == nil {
		translogMutex.Lock()

		if translogInstance == nil {
			translogInstance = new(transLogBuffer)
		}

		translogMutex.Unlock()
	}

	return translogInstance
}

func (tb *transLogBuffer) AddToBuffer(message string) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	_, err := tb.buffer.WriteString(message)

	if err != nil {
		log.Printf("TLB-Error:%v\n", err)
	}
}

func (tb *transLogBuffer) Flush() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	filename := tb.getTranslogFilePath()
	if tb.buffer.Len() == 0 {
		return
	}

	tb.flushBufferToFile(&tb.buffer, filename)
	tb.buffer.Reset()
}

func ReplayTranslog(forceReply bool) (int64, error) {
	var keycount int64 = 0

	filepath := config.GetTransactionLogFilePath()
	filePtr, _ := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)
	defer filePtr.Close()

	buffer := bufio.NewReader(filePtr)

	for {
		decoded, err := resp3.Decode(buffer)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return keycount, fmt.Errorf("failed to parse a commands from translog, "+
					"please fix the translog file, or run the server with force replay enabled "+
					"in config to skip the record and continue further. ERR=[%v]", err)
			}
		}

		command, cmderr := getCommandFromRESP(decoded)
		if cmderr != nil {
			if forceReply {
				continue
			}

			return keycount, fmt.Errorf("failed to parse a commands from translog, "+
				"please fix the translog file, or run the server with force replay enabled "+
				"in config to skip the record and continue further. ERR=[%v]", err)
		}

		_, execErr := executeCommand(command)
		if execErr != nil {
			if forceReply {
				continue
			}

			return keycount, errors.New("failed to replay a commands into the memorystore, " +
				"potentially errornous translog or intermittent write failure. Run the server " +
				"with force replay enabled in config to skip the record and continue further")
		}

		log.Printf(" >> Reply Done for command: %d\n", keycount)
		keycount++
	}

	filePtr.Truncate(0) // truncate the translog for fresh records
	return keycount, nil
}

func (tb *transLogBuffer) flushBufferToFile(buffer *bytes.Buffer, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

func (tb *transLogBuffer) getTranslogFilePath() string {
	return config.GetTransactionLogFilePath()
}
