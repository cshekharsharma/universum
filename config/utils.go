package config

import "time"

func GetServerPort() int64 {
	port, err := GetInt64("ServerPort", APP_CODE_NAME)

	if err != nil {
		port = DEFAULT_SERVER_PORT
	}

	return port
}

func GetMaxClientConnections() int64 {
	maxClients, err := GetInt64("MaxConnections", APP_CODE_NAME)

	if err != nil {
		maxClients = MAX_CLIENT_CONNECTIONS
	}

	if maxClients < 1 || maxClients > MAX_CLIENT_CONNECTIONS {
		maxClients = MAX_CLIENT_CONNECTIONS
	}

	return maxClients
}

func GetRequestExecutionTimeout() time.Duration {
	timeout, err := GetInt64("RequestExecutionTimeout", APP_CODE_NAME)

	if err != nil {
		timeout = DEFAULT_REQUEST_EXEC_TIMEOUT
	}

	return time.Duration(timeout) * time.Second
}

func GetTCPConnectionWriteTimeout() time.Duration {
	timeout, err := GetInt64("ConnectionWriteTimeout", APP_CODE_NAME)

	if err != nil {
		timeout = DEFAULT_CONN_WRITE_TIMEOUT
	}

	return time.Duration(timeout) * time.Second
}

func GetAllowedMemoryStorageLimit() int64 {
	limit, err := GetInt64("AllowedMemoryStorageLimit", APP_CODE_NAME)

	if err != nil {
		limit = 0
	}

	return limit
}

func GetTransactionLogFilePath() string {
	path, err := GetString("TransactionLogFilePath", APP_CODE_NAME)

	if err != nil {
		path = DEFAULT_TRANSLOG_FILE_PATH
	}

	return path
}

func GetServerLogFilePath() string {
	path, err := GetString("ServerLogFilePath", APP_CODE_NAME)

	if err != nil {
		path = DEFAULT_SERVER_LOG_FILE_PATH
	}

	return path
}

func GetAutoRecordExpiryFrequency() time.Duration {
	frequency, err := GetInt64("AutoRecordExpiryFrequency", APP_CODE_NAME)

	if err != nil {
		frequency = DEFAULT_AUTO_EXPIRY_FREQUENCY
	}

	return time.Duration(frequency) * time.Second
}

func GetAutoSnapshotFrequency() time.Duration {
	frequency, err := GetInt64("AutoSnapshotFrequency", APP_CODE_NAME)

	if err != nil {
		frequency = DEFAULT_AUTO_SNAPSHOT_FREQUENCY
	}

	return time.Duration(frequency) * time.Second
}

func GetRecordAutoEvictionPolicy() string {
	policy, err := GetString("RecordAutoEvictionPolicy", APP_CODE_NAME)

	if err != nil {
		policy = DEFAULT_AUTO_EVICTION_POLICY
	}

	return policy
}

func GetMinimumLogLevel() string {
	level, err := GetString("MinimumLogLevel", APP_CODE_NAME)

	if err != nil {
		level = DEFAULT_MINIMUM_LOG_LEVEL
	}

	return level
}
