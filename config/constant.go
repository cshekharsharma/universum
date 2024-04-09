package config

const APP_NAME_LABEL string = "UniversumDB"
const APP_CODE_NAME string = "universum"

const DEFAULT_CONFIG_NAME string = "config.ini"

const DEFAULT_SERVER_PORT int64 = 11191

const MAX_CLIENT_CONNECTIONS int64 = 100000
const MAX_SERVER_CONCURRENCY int64 = 600
const DEFAULT_CONN_READ_TIMEOUT int64 = 10

const DEFAULT_TRANSLOG_FILE_PATH string = "/opt/universum/translog.aof"
const DEFAULT_SERVER_LOG_FILE_PATH string = "/var/log/universum/server.log"
