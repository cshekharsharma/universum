package config

import "time"

// Section:Server
func GetServerPort() int64 {
	port, err := GetInt64("ServerPort", SectionServer)
	if err != nil {
		port = DefaultServerPort
	}
	return port
}

func GetMaxClientConnections() int64 {
	maxClients, err := GetInt64("MaxConnections", SectionServer)
	if err != nil {
		maxClients = MaxClientConnections
	}

	if maxClients < 1 || maxClients > MaxClientConnections {
		maxClients = MaxClientConnections
	}

	return maxClients
}

func GetRequestExecutionTimeout() time.Duration {
	timeout, err := GetInt64("RequestExecutionTimeout", SectionServer)
	if err != nil {
		timeout = DefaultRequestExecTimeout
	}
	return time.Duration(timeout) * time.Second
}

func GetTCPConnectionWriteTimeout() time.Duration {
	timeout, err := GetInt64("ConnectionWriteTimeout", SectionServer)
	if err != nil {
		timeout = DefaultConnectionWriteTimeout
	}
	return time.Duration(timeout) * time.Second
}

// Section:Storage
func GetStorageEngineType() string {
	store, err := GetString("StorageEngine", SectionStorage)
	if err != nil {
		store = DefaultStorageEngine
	}
	return store
}

func GetMaxRecordSizeInBytes() int64 {
	size, err := GetInt64("MaxRecordSizeInBytes", SectionStorage)
	if err != nil {
		size = DefaultMaxRecordSizeInBytes
	}

	if size > DefaultMaxRecordSizeInBytes {
		return DefaultMaxRecordSizeInBytes
	}

	return size
}

// Section:Storage.Memory
func GetAllowedMemoryStorageLimit() int64 {
	limit, err := GetInt64("AllowedMemoryStorageLimit", SectionStorageMemory)
	if err != nil {
		limit = 0
	}
	return limit
}

func GetAutoSnapshotFrequency() time.Duration {
	frequency, err := GetInt64("AutoSnapshotFrequency", SectionStorageMemory)
	if err != nil {
		frequency = DefaultAutoSnapshotFrequency
	}
	return time.Duration(frequency) * time.Second
}

func GetSnapshotFileDirectory() string {
	path, err := GetString("SnapshotFileDirectory", SectionStorageMemory)
	if err != nil {
		path = DefaultSnapshotFileDirectory
	}
	return path
}

// Section:Storage.LSM
func GetMemtableStorageType() string {
	mtype, err := GetString("MemtableStorageType", SectionStorageLSM)
	if err != nil {
		mtype = MemtableStorageTypeLB
	}
	return mtype
}

func GetLSMWriteBlockSize() int64 {
	blockSize, err := GetInt64("DefaultWriteBlockSize", SectionStorageLSM)
	if err != nil {
		blockSize = DefaultWriteBlockSize
	}
	return blockSize
}

func GetDataStorageDirectory() string {
	path, err := GetString("DataStorageDirectory", SectionStorageLSM)
	if err != nil {
		path = DefaultDataStorageDirectory
	}
	return path
}

func GetWriteAheadLogDirectory() string {
	path, err := GetString("WriteAheadLogDirectory", SectionStorageLSM)
	if err != nil {
		path = DefaultWriteAheadLogDirectory
	}
	return path
}

func GetWriteAheadLogFrequency() time.Duration {
	frequency, err := GetInt64("WriteAheadLogFrequency", SectionStorageLSM)
	if err != nil {
		frequency = DefaultWriteAheadLogFrequency
	}
	return time.Duration(frequency) * time.Second
}

func GetWriteAheadLogBufferSize() int64 {
	size, err := GetInt64("WriteAheadLogBufferSize", SectionStorageLSM)
	if err != nil {
		size = DefaultWriteAheadLogBufferSize
	}
	return size
}

// Section:Eviction
func GetAutoRecordExpiryFrequency() time.Duration {
	frequency, err := GetInt64("AutoRecordExpiryFrequency", SectionEviction)
	if err != nil {
		frequency = DefaultAutoExpiryFrequency
	}
	return time.Duration(frequency) * time.Second
}

func GetRecordAutoEvictionPolicy() string {
	policy, err := GetString("RecordAutoEvictionPolicy", SectionEviction)
	if err != nil {
		policy = DefaultAutoEvictionPolicy
	}
	return policy
}

// Section:Logging
func GetServerLogFilePath() string {
	path, err := GetString("ServerLogFilePath", SectionLogging)
	if err != nil {
		path = DefaultServerLogFilePath
	}
	return path
}

func GetMinimumLogLevel() string {
	level, err := GetString("MinimumLogLevel", SectionLogging)

	if err != nil {
		level = DefaultMinimumLogLevel
	}

	return level
}

// Section:Auth
func IsAuthenticationEnabled() bool {
	enabled, err := GetBool("AuthenticationEnabled", SectionAuth)
	if err != nil {
		return false
	}
	return enabled
}

func GetDbUserName() string {
	user, err := GetString("DbUserName", SectionAuth)
	if err != nil {
		return ""
	}
	return user
}

func GetDbUserPassword() string {
	pass, err := GetString("DbUserPassword", SectionAuth)
	if err != nil {
		return ""
	}
	return pass
}
