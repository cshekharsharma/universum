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

	// Constants for valid field values
	StorageTypeMemory     string = "MEMORY"
	StorageTypeLSM        string = "LSM"
	MemtableStorageTypeLB string = "LB" // skip list + bloom filter
	MemtableStorageTypeLM string = "LM" // sorted list + sync.Map

	// Config section names
	SectionServer        string = "server"
	SectionLogging       string = "logging"
	SectionSnapshot      string = "snapshot"
	SectionStorage       string = "storage"
	SectionAuth          string = "auth"
	SectionEviction      string = "eviction"
	SectionStorageMemory string = "storage.memory"
	SectionStorageLSM    string = "storage.lsm"

	// Section:Server
	DefaultServerPort             int64 = 11191
	MaxClientConnections          int64 = 100000
	DefaultConnectionWriteTimeout int64 = 10
	DefaultRequestExecTimeout     int64 = 10

	// Storage
	DefaultStorageEngine        string = "MEMORY"
	DefaultMaxRecordSizeInBytes int64  = 1024 * 1024 // 1 MB

	// Storage.Memory
	AllowedMemoryStorageLimit    int64  = 1024 * 1024 * 1024 // 1 GB
	DefaultAutoSnapshotFrequency int64  = 10
	DefaultSnapshotFileDirectory string = "/opt/universum/snapshot"

	// Storage.LSM
	DefaultMemtableStorageType  string = MemtableStorageTypeLB
	DefaultWriteBlockSize       int64  = 65536
	DefaultDataStorageDirectory string = "/opt/universum/data"
	DefaultAOFTranslogDirectory string = "/opt/universum/translog"

	// Section:Logging
	DefaultServerLogFilePath string = "/var/log/universum/server.log"
	DefaultMinimumLogLevel   string = "INFO"

	// Section:Eviction
	DefaultAutoExpiryFrequency int64  = 2
	DefaultAutoEvictionPolicy  string = "NONE"

	// Section:Auth
	DefaultAuthenticationMode int64 = 1
)
