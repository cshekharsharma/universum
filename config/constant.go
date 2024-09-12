package config

var (
	AppVersion string
	BuildTime  int64
	GitHash    string
	AppEnv     string
)

const (
	APP_NAME_LABEL string = "UniversumDB"
	APP_CODE_NAME  string = "universum"

	DEFAULT_CONFIG_NAME string = "config.ini"

	DEFAULT_SERVER_PORT int64 = 11191

	MAX_CLIENT_CONNECTIONS          int64  = 100000
	MAX_SERVER_CONCURRENCY          int64  = 600
	DEFAULT_REQUEST_EXEC_TIMEOUT    int64  = 10
	DEFAULT_CONN_WRITE_TIMEOUT      int64  = 10
	DEFAULT_AUTO_EXPIRY_FREQUENCY   int64  = 2
	DEFAULT_AUTO_SNAPSHOT_FREQUENCY int64  = 10
	DEFAULT_AUTO_EVICTION_POLICY    string = "NONE"

	DEFAULT_TRANSLOG_FILE_PATH   string = "/opt/universum/translog.aof"
	DEFAULT_SERVER_LOG_FILE_PATH string = "/var/log/universum/server.log"
)
