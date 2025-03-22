package config

import (
	"errors"
	"fmt"
	"strings"
	"universum/utils"
	"universum/utils/filesys"
)

// Config validator
type ConfigValidator struct {
}

func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

func (v *ConfigValidator) Validate(config *Config) (*Config, error) {

	err := v.validateServerSection(config)
	if err != nil {
		return nil, err
	}

	err = v.validateClusterSection(config)
	if err != nil {
		return nil, err
	}

	err = v.validateLoggingSection(config)
	if err != nil {
		return nil, err
	}

	err = v.validateStorageSection(config)
	if err != nil {
		return nil, err
	}
	err = v.validateEvictionSection(config)
	if err != nil {
		return nil, err
	}

	err = v.validateAuthSection(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (v *ConfigValidator) validateServerSection(config *Config) error {
	if config.Server == nil {
		return fmt.Errorf("server section is missing in config")
	}

	if config.Server.ServerPort == 0 {
		config.Server.ServerPort = DefaultServerPort
	}

	if config.Server.ServerPort < 1 || config.Server.ServerPort > 65535 {
		return fmt.Errorf("invalid port number %d", config.Server.ServerPort)
	}

	if config.Server.ConnectionWriteTimeout == 0 {
		config.Server.ConnectionWriteTimeout = DefaultConnectionWriteTimeout
	}

	if config.Server.RequestExecutionTimeout == 0 {
		config.Server.RequestExecutionTimeout = DefaultRequestExecTimeout
	}

	if config.Server.MaxConnections == 0 {
		config.Server.MaxConnections = DefultMaxClientConnections
	}

	if config.Server.EnableTLS && config.Server.TLSCertFilePath == "" {
		config.Server.TLSCertFilePath = DefaultTLSCertFilePath
	}

	if config.Server.EnableTLS && config.Server.TLSKeyFilePath == "" {
		config.Server.TLSKeyFilePath = DefaultTLSKeyFilePath
	}

	return nil
}

func (v *ConfigValidator) validateClusterSection(config *Config) error {
	if config.Cluster == nil {
		return errors.New("cluster section is missing in config")
	}

	if config.Cluster.HeartbeatPort == 0 {
		config.Cluster.HeartbeatPort = DefaultHeartbeatPort
	}

	if config.Cluster.HeartbeatPort == config.Server.ServerPort {
		return fmt.Errorf("heartbeat port %d cannot be same as server port %d",
			config.Cluster.HeartbeatPort, config.Server.ServerPort)
	}

	if config.Cluster.HeartbeatIntervalMs == 0 {
		config.Cluster.HeartbeatIntervalMs = DefaultHeartbeatIntervalMs
	}

	if config.Cluster.EnableCluster && len(config.Cluster.Hosts) < int(MinClusterSize) {
		return fmt.Errorf("host list cannot be less than %d in cluster mode [potential split brain]", MinClusterSize)
	}

	if config.Cluster.ReplicationFactor <= 0 {
		config.Cluster.ReplicationFactor = DefaultReplicationFactor
	}

	return nil
}

func (v *ConfigValidator) validateStorageSection(config *Config) error {
	if config.Storage == nil {
		return errors.New("storage section is missing in config")
	}

	if config.Storage.StorageEngine == "" {
		config.Storage.StorageEngine = DefaultStorageEngine
	}

	allowedStorageEngines := []string{StorageEngineLSM, StorageEngineMemory}
	config.Storage.StorageEngine = strings.ToUpper(config.Storage.StorageEngine)

	if exists, _ := utils.ExistsInList(config.Storage.StorageEngine, allowedStorageEngines); !exists {
		return fmt.Errorf("invalid storage engine %s set in config", config.Storage.StorageEngine)
	}

	if config.Storage.MaxRecordSizeInBytes == 0 {
		config.Storage.MaxRecordSizeInBytes = DefaultMaxRecordSizeInBytes
	}

	if config.Storage.StorageEngine == StorageEngineLSM {
		err := v.validateStorageEngineLSM(config)
		if err != nil {
			return err
		}
	}

	if config.Storage.StorageEngine == StorageEngineMemory {
		err := v.validateStorageEngineMemory(config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *ConfigValidator) validateStorageEngineLSM(config *Config) error {
	if config.Storage.LSM == nil {
		return errors.New("LSM section cannot be empty for LSM storage engine")
	}

	if config.Storage.LSM.MemtableStorageType == "" {
		config.Storage.LSM.MemtableStorageType = DefaultMemtableStorageType
	}

	allowedMemtableStorageTypes := []string{MemtableStorageTypeLB}
	config.Storage.LSM.MemtableStorageType = strings.ToUpper(config.Storage.LSM.MemtableStorageType)

	if exists, _ := utils.ExistsInList(config.Storage.LSM.MemtableStorageType, allowedMemtableStorageTypes); !exists {
		return fmt.Errorf("invalid memtable storage type %s set in config", config.Storage.LSM.MemtableStorageType)
	}

	if config.Storage.LSM.BloomFilterMaxRecords == 0 {
		config.Storage.LSM.BloomFilterMaxRecords = DefaultBloomFilterMaxRecords
	}

	if config.Storage.LSM.WriteBufferSize == 0 {
		config.Storage.LSM.WriteBufferSize = DefaultWriteBlockSize
	}

	if config.Storage.LSM.BloomFalsePositiveRate == 0 {
		config.Storage.LSM.BloomFalsePositiveRate = DefaultBloomFalsePositiveRate
	}

	if config.Storage.LSM.WriteBlockSize == 0 {
		config.Storage.LSM.WriteBlockSize = DefaultWriteBlockSize
	}

	if config.Storage.LSM.BlockCompressionAlgo == "" {
		config.Storage.LSM.BlockCompressionAlgo = DefaultBlockCompressionAlgo
	}

	config.Storage.LSM.BlockCompressionAlgo = strings.ToUpper(config.Storage.LSM.BlockCompressionAlgo)
	if exists, _ := utils.ExistsInList(config.Storage.LSM.BlockCompressionAlgo, AllowedCompressionAlgos); !exists {
		return fmt.Errorf("invalid block compression algo %s set in config", config.Storage.LSM.BlockCompressionAlgo)
	}

	if config.Storage.LSM.DataStorageDirectory == "" {
		config.Storage.LSM.DataStorageDirectory = DefaultDataStorageDirectory
	}

	if config.Storage.LSM.WriteAheadLogBufferSize == 0 {
		config.Storage.LSM.WriteAheadLogBufferSize = DefaultWriteAheadLogBufferSize
	}

	if config.Storage.LSM.WriteAheadLogFrequency == 0 {
		config.Storage.LSM.WriteAheadLogFrequency = DefaultWriteAheadLogFrequency
	}

	if config.Storage.LSM.WriteAheadLogDirectory == "" {
		config.Storage.LSM.WriteAheadLogDirectory = DefaultWriteAheadLogDirectory
	}

	if !filesys.IsDirectoryWritable(config.Storage.LSM.WriteAheadLogDirectory) {
		return fmt.Errorf("WAL directory %s is not writable", config.Storage.LSM.WriteAheadLogDirectory)
	}

	if !filesys.IsDirectoryWritable(config.Storage.LSM.DataStorageDirectory) {
		return fmt.Errorf("data directory %s is not writable", config.Storage.LSM.DataStorageDirectory)
	}

	if config.Storage.LSM.WriteAheadLogDirectory == "" {
		config.Storage.LSM.WriteAheadLogDirectory = DefaultWriteAheadLogDirectory
	}

	if !filesys.IsDirectoryWritable(config.Storage.LSM.WriteAheadLogDirectory) {
		return fmt.Errorf("WAL directory %s is not writable", config.Storage.LSM.WriteAheadLogDirectory)
	}

	if config.Storage.LSM.BlockCacheMemoryLimit == 0 {
		config.Storage.LSM.BlockCacheMemoryLimit = DefaultBlockCacheMemoryLimit
	}

	return nil
}
func (v *ConfigValidator) validateStorageEngineMemory(config *Config) error {
	if config.Storage.Memory == nil {
		return errors.New("memory section cannot be empty for memory storage engine")
	}

	if config.Storage.Memory.AllowedMemoryStorageLimit == 0 {
		config.Storage.Memory.AllowedMemoryStorageLimit = AllowedMemoryStorageLimit
	}

	if config.Storage.Memory.SnapshotFileDirectory == "" {
		config.Storage.Memory.SnapshotFileDirectory = DefaultSnapshotFileDirectory
	}

	if !filesys.IsDirectoryWritable(config.Storage.Memory.SnapshotFileDirectory) {
		return fmt.Errorf("snapshot file directory %s is not writable", config.Storage.Memory.SnapshotFileDirectory)
	}

	if config.Storage.Memory.AutoSnapshotFrequency == 0 {
		config.Storage.Memory.AutoSnapshotFrequency = DefaultAutoSnapshotFrequency
	}

	if config.Storage.Memory.SnapshotCompressionAlgo == "" {
		config.Storage.Memory.SnapshotCompressionAlgo = DefaultSnapshotCompressionAlgo
	}

	config.Storage.Memory.SnapshotCompressionAlgo = strings.ToUpper(config.Storage.Memory.SnapshotCompressionAlgo)
	if exists, _ := utils.ExistsInList(config.Storage.Memory.SnapshotCompressionAlgo, AllowedCompressionAlgos); !exists {
		return fmt.Errorf("invalid snapshotcompression algo %s set in config", config.Storage.Memory.SnapshotCompressionAlgo)
	}

	return nil
}

func (v *ConfigValidator) validateEvictionSection(config *Config) error {
	if config.Eviction == nil {
		return errors.New("eviction section is missing in config")
	}

	if config.Eviction.AutoEvictionPolicy == "" {
		config.Eviction.AutoEvictionPolicy = DefaultAutoEvictionPolicy
	}

	config.Eviction.AutoEvictionPolicy = strings.ToUpper(config.Eviction.AutoEvictionPolicy)
	if exists, _ := utils.ExistsInList(config.Eviction.AutoEvictionPolicy, AllowedAutoEvictionPolicy); !exists {
		return fmt.Errorf("invalid eviction policy %s set in config", config.Eviction.AutoEvictionPolicy)
	}

	if config.Eviction.AutoRecordExpiryFrequency == 0 {
		config.Eviction.AutoRecordExpiryFrequency = DefaultRecordAutoExpiryFrequency
	}

	return nil
}

func (v *ConfigValidator) validateLoggingSection(config *Config) error {
	if config.Logging == nil {
		return errors.New("logging section is missing in config")
	}

	if config.Logging.MinimumLogLevel == "" {
		config.Logging.MinimumLogLevel = DefaultMinimumLogLevel
	}

	config.Logging.MinimumLogLevel = strings.ToUpper(config.Logging.MinimumLogLevel)
	if exists, _ := utils.ExistsInList(config.Logging.MinimumLogLevel, AllowedLogLevels); !exists {
		return fmt.Errorf("invalid log level %s set in config", config.Logging.MinimumLogLevel)
	}

	if config.Logging.LogFileDirectory == "" {
		config.Logging.LogFileDirectory = DefaultLogFileDirectory
	}

	if !filesys.IsDirectoryWritable(config.Logging.LogFileDirectory) {
		return fmt.Errorf("log file directory %s is not writable", config.Logging.LogFileDirectory)
	}

	return nil
}

func (v *ConfigValidator) validateAuthSection(config *Config) error {
	if config.Auth == nil {
		return errors.New("auth section is missing in config")
	}
	return nil
}
