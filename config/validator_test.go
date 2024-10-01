package config

import (
	"testing"
)

func TestConfigValidator_Validate(t *testing.T) {
	validator := NewConfigValidator()

	t.Run("ValidateServerSection", func(t *testing.T) {
		cfg := GetSkeleton()
		cfg.Server = &Server{
			ServerPort:              0,
			MaxConnections:          0,
			ConnectionWriteTimeout:  30,
			RequestExecutionTimeout: 0,
		}

		err := validator.validateServerSection(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.Server.ServerPort != DefaultServerPort {
			t.Errorf("Expected ServerPort to be set to %d, got %d", DefaultServerPort, cfg.Server.ServerPort)
		}

		if cfg.Server.MaxConnections != DefultMaxClientConnections {
			t.Errorf("Expected MaxConnections to be set to %d, got %d", DefultMaxClientConnections, cfg.Server.MaxConnections)
		}

		if cfg.Server.ConnectionWriteTimeout != 30 {
			t.Errorf("Expected ConnectionWriteTimeout to be set to 30, got %d", cfg.Server.ConnectionWriteTimeout)
		}
	})

	t.Run("ValidateLoggingSection", func(t *testing.T) {
		cfg := GetSkeleton()
		cfg.Logging = &Logging{
			MinimumLogLevel:  "",
			LogFileDirectory: "/invalid/log/dir",
		}

		err := validator.validateLoggingSection(cfg)
		if err == nil {
			t.Fatalf("Expected error, got %v", err)
		}

		if cfg.Logging.MinimumLogLevel != DefaultMinimumLogLevel {
			t.Errorf("Expected MinimumLogLevel to be set to %s, got %s", DefaultMinimumLogLevel, cfg.Logging.MinimumLogLevel)
		}
	})

	t.Run("ValidateStorageSection", func(t *testing.T) {
		cfg := GetSkeleton()

		cfg.Storage = &Storage{
			StorageEngine:        "invalid-engine",
			MaxRecordSizeInBytes: 0,
			Memory: &Memory{
				SnapshotFileDirectory: "/tmp",
			},
		}

		err := validator.validateStorageSection(cfg)
		if err == nil {
			t.Fatalf("Expected an error for invalid storage engine, but got none")
		}

		cfg.Storage = &Storage{
			StorageEngine:        "",
			MaxRecordSizeInBytes: 0,
			Memory: &Memory{
				SnapshotFileDirectory: "/tmp",
			},
		}

		err = validator.validateStorageSection(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.Storage.StorageEngine != DefaultStorageEngine {
			t.Errorf("Expected StorageEngine to be set to %s, got %s", DefaultStorageEngine, cfg.Storage.StorageEngine)
		}

		if cfg.Storage.MaxRecordSizeInBytes != DefaultMaxRecordSizeInBytes {
			t.Errorf("Expected MaxRecordSizeInBytes to be set to %d, got %d", DefaultMaxRecordSizeInBytes, cfg.Storage.MaxRecordSizeInBytes)
		}
	})

	t.Run("ValidateStorageSectionWithLSMEngine", func(t *testing.T) {
		cfg := GetSkeleton()
		cfg.Storage.StorageEngine = StorageEngineLSM

		cfg.Storage.LSM = &LSM{
			WriteBlockSize:          0,
			MemtableStorageType:     "",
			DataStorageDirectory:    "/tmp",
			WriteAheadLogDirectory:  "/tmp",
			WriteAheadLogFrequency:  0,
			WriteAheadLogBufferSize: 1024,
		}

		err := validator.validateStorageEngineLSM(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.Storage.LSM.WriteBlockSize != DefaultWriteBlockSize {
			t.Errorf("Expected WriteBlockSize to be set to %d, got %d", DefaultWriteBlockSize, cfg.Storage.LSM.WriteBlockSize)
		}

		if cfg.Storage.LSM.MemtableStorageType != DefaultMemtableStorageType {
			t.Errorf("Expected MemtableStorageType to be set to %s, got %s", DefaultMemtableStorageType, cfg.Storage.LSM.MemtableStorageType)
		}

		if cfg.Storage.LSM.WriteAheadLogFrequency != DefaultWriteAheadLogFrequency {
			t.Errorf("Expected WriteAheadLogFrequency to be set to %d, got %d", DefaultWriteAheadLogFrequency, cfg.Storage.LSM.WriteAheadLogFrequency)
		}

		if cfg.Storage.LSM.WriteAheadLogBufferSize != 1024 {
			t.Errorf("Expected WriteAheadLogBufferSize to be set to 1024, got %d", cfg.Storage.LSM.WriteAheadLogBufferSize)
		}

		if cfg.Storage.LSM.DataStorageDirectory != "/tmp" {
			t.Errorf("Expected DataStorageDirectory to be set to /tmp, got %s", cfg.Storage.LSM.DataStorageDirectory)
		}

		if cfg.Storage.LSM.WriteAheadLogDirectory != "/tmp" {
			t.Errorf("Expected WriteAheadLogDirectory to be set to /tmp, got %s", cfg.Storage.LSM.WriteAheadLogDirectory)
		}
	})

	t.Run("ValidateStorageSectionWithMemoryEngine", func(t *testing.T) {
		cfg := GetSkeleton()
		cfg.Storage.StorageEngine = StorageEngineMemory

		cfg.Storage.Memory = &Memory{
			AllowedMemoryStorageLimit: 0,
			SnapshotFileDirectory:     "/tmp",
			AutoSnapshotFrequency:     100,
		}

		err := validator.validateStorageEngineMemory(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.Storage.Memory.AllowedMemoryStorageLimit != AllowedMemoryStorageLimit {
			t.Errorf("Expected AllowedMemoryStorageLimit to be set to %d, got %d", AllowedMemoryStorageLimit, cfg.Storage.Memory.AllowedMemoryStorageLimit)
		}

		if cfg.Storage.Memory.AutoSnapshotFrequency != 100 {
			t.Errorf("Expected AutoSnapshotFrequency to be set to 100, got %d", cfg.Storage.Memory.AllowedMemoryStorageLimit)
		}
	})

	t.Run("validateEvictionSection", func(t *testing.T) {
		cfg := GetSkeleton()
		cfg.Eviction = &Eviction{
			AutoEvictionPolicy:        "",
			AutoRecordExpiryFrequency: 0,
		}

		err := validator.validateEvictionSection(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.Eviction.AutoEvictionPolicy != DefaultAutoEvictionPolicy {
			t.Errorf("Expected AutoEvictionPolicy to be set to %s, got %s", DefaultAutoEvictionPolicy, cfg.Eviction.AutoEvictionPolicy)
		}
		if cfg.Eviction.AutoRecordExpiryFrequency != DefaultRecordAutoExpiryFrequency {
			t.Errorf("Expected AutoRecordExpiryFrequency to be set to %d, got %d", DefaultRecordAutoExpiryFrequency, cfg.Eviction.AutoRecordExpiryFrequency)
		}
	})
}
