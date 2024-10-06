package wal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"universum/config"
	"universum/entity"
	"universum/internal/logger"
	"universum/resp3"
	"universum/storage/lsm/memtable"
	"universum/utils"
)

type WALReader struct {
	fileptr *os.File
}

func NewReader(filedir string) (*WALReader, error) {
	filePath := filepath.Clean(filepath.Join(filedir, config.DefaultWALFileName))
	fileptr, err := os.Open(filePath)

	if err != nil {
		return nil, fmt.Errorf("WALReader: failed to open WAL file: %v", err)
	}

	return &WALReader{
		fileptr: fileptr,
	}, nil
}

func (wr *WALReader) readEntries() ([]*WALRecord, error) {
	var entries []*WALRecord

	for {
		var commandLen int64
		err := binary.Read(wr.fileptr, binary.BigEndian, &commandLen)
		if err == io.EOF {
			break // Reached end of WAL file
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read command length: %v", err)
		}

		commandBytes := make([]byte, commandLen)
		_, err = wr.fileptr.Read(commandBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read command bytes: %v", err)
		}

		buf := bytes.NewReader(commandBytes)
		reader := bufio.NewReader(buf)
		command, err := resp3.Decode(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decode command: %v", err)
		}

		parsedCommand, ok := command.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to parse decoded command as map")
		}

		entry := &WALRecord{
			Operation: parsedCommand["Name"].(string),
			Key:       parsedCommand["Key"].(string),
		}
		if entry.Operation == OperationTypeSET {
			entry.Value = parsedCommand["Value"]
			entry.TTL = parsedCommand["Expiry"].(int64)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (wr *WALReader) RestoreFromWAL(memTable memtable.MemTable) error {
	entries, err := wr.readEntries()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		switch entry.Operation {
		case OperationTypeSET:
			entry.TTL = entry.TTL - utils.GetCurrentEPochTime()

			didSet, code := memTable.Set(entry.Key, entry.Value, entry.TTL)
			if !didSet && code != entity.CRC_RECORD_UPDATED {
				logger.Get().Warn("failed to restore record key=%s from WAL: %v", entry.Key, code)
			}

		case OperationTypeDELETE:
			deleted, code := memTable.Delete(entry.Key)
			if !deleted && code != entity.CRC_RECORD_DELETED {
				logger.Get().Warn("failed to restore record key=%s from WAL: %v", entry.Key, code)
			}

		default:
			return fmt.Errorf("unknown operation type in WAL entry: %s", entry.Operation)
		}
	}

	return nil
}

func (wr *WALReader) Close() {
	wr.fileptr.Close()
}