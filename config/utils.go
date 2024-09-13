package config

import "time"

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
		timeout = DefaultConnWriteTimeout
	}
	return time.Duration(timeout) * time.Second
}

// storage configs
func GetStorageEngine() string {
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

func GetAllowedMemoryStorageLimit() int64 {
	limit, err := GetInt64("AllowedMemoryStorageLimit", SectionStorage)
	if err != nil {
		limit = 0
	}
	return limit
}

// Snapshot configs
func GetTransactionLogFilePath() string {
	path, err := GetString("TransactionLogFilePath", SectionSnapshot)
	if err != nil {
		path = DefaultTranslogFilePath
	}
	return path
}

func GetAutoSnapshotFrequency() time.Duration {
	frequency, err := GetInt64("AutoSnapshotFrequency", SectionSnapshot)
	if err != nil {
		frequency = DefaultAutoSnapshotFrequency
	}
	return time.Duration(frequency) * time.Second
}

// Eviction configs
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

// Logging configs
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

// Auth configs
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
