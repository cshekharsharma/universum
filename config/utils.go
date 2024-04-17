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

func GetServerConcurrencyLimit(maxConnections int64) int64 {
	concurrency, err := GetInt64("MaxConcurrency", APP_CODE_NAME)

	if err != nil {
		concurrency = MAX_SERVER_CONCURRENCY
	}

	if concurrency < 1 {
		concurrency = MAX_SERVER_CONCURRENCY
	}

	// making sure that concurrency level is not more than maximum
	// allowed connection limit. Since that will just be a waste of
	// resources, we will never reach 100% utilisation concurrency.
	if concurrency > maxConnections {
		concurrency = int64(maxConnections)
	}

	return concurrency
}

func GetTCPConnectionReadtime() time.Duration {
	timeout, err := GetInt64("ConnectionReadTimeout", APP_CODE_NAME)

	if err != nil {
		timeout = DEFAULT_CONN_READ_TIMEOUT
	}

	return time.Duration(timeout) * time.Minute
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
