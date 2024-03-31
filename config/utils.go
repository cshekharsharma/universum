package config

import "time"

func GetServerPort() int {
	port, err := GetInt64("ServerPort", APP_CODE_NAME)

	if err != nil {
		port = DEFAULT_SERVER_PORT
	}

	return int(port)
}

func GetMaxClientConnections() int {
	maxClients, err := GetInt64("MaxConnections", APP_CODE_NAME)

	if err != nil {
		maxClients = MAX_CLIENT_CONNECTIONS
	}

	if maxClients < 1 || maxClients > MAX_CLIENT_CONNECTIONS {
		maxClients = MAX_CLIENT_CONNECTIONS
	}

	return int(maxClients)
}

func GetServerConcurrencyLimit(maxConnections int) int {
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
	if int(concurrency) > maxConnections {
		concurrency = int64(maxConnections)
	}

	return int(concurrency)
}

func GetTCPConnectionReadtime() time.Duration {
	timeout, err := GetInt64("ConnectionReadTimeout", APP_CODE_NAME)

	if err != nil {
		timeout = DEFAULT_CONN_READ_TIMEOUT
	}

	return time.Duration(timeout) * time.Minute
}
