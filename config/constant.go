package config

var (
	AppVersion string
	BuildTime  int64
	GitHash    string
	AppEnv     string
)

const (
	AppNameLabel string = "UniversumDB"
	AppCodeName  string = "universum"

	DefaultConfigName string = "config.ini"

	// Config section names
	SectionServer        string = "server"
	SectionLogging       string = "logging"
	SectionSnapshot      string = "snapshot"
	SectionStorage       string = "storage"
	SectionAuth          string = "auth"
	SectionEviction      string = "eviction"
	SectionStorageMemory string = "storage.memory"

	// Default configuration values
	DefaultServerPort            int64  = 11191
	MaxClientConnections         int64  = 100000
	MaxServerConcurrency         int64  = 600
	DefaultRequestExecTimeout    int64  = 10
	DefaultConnWriteTimeout      int64  = 10
	DefaultAutoExpiryFrequency   int64  = 2
	DefaultAutoSnapshotFrequency int64  = 10
	DefaultAutoEvictionPolicy    string = "NONE"
	DefaultMinimumLogLevel       string = "INFO"
	DefaultStorageEngine         string = "MEMORY"
	DefaultMaxRecordSizeInBytes  int64  = 1024 * 1024

	DefaultTranslogFilePath  string = "/opt/universum/translog.aof"
	DefaultServerLogFilePath string = "/var/log/universum/server.log"

	StorageTypeMemory = "MEMORY"
)
