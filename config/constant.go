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

	DefaultConfigName  string = "config.ini"
	DefaultWALFileName string = "writeahead.aof"

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
	MaxClientConnections          int64 = 10000
	DefaultConnectionWriteTimeout int64 = 10 // 10 seconds
	DefaultRequestExecTimeout     int64 = 10 // 10 seconds

	// Storage
	DefaultStorageEngine        string = "MEMORY"    // [MEMORY, LSM]
	DefaultMaxRecordSizeInBytes int64  = 1024 * 1024 // 1 MB

	// Storage.Memory
	AllowedMemoryStorageLimit    int64  = 1024 * 1024 * 1024 // 1 GB
	DefaultAutoSnapshotFrequency int64  = 10                 // 10 seconds
	DefaultSnapshotFileDirectory string = "/opt/universum/snapshot"

	// Storage.LSM
	DefaultMemtableStorageType     string = MemtableStorageTypeLB
	DefaultWriteBlockSize          int64  = 65536 // 64 KB
	DefaultDataStorageDirectory    string = "/opt/universum/data"
	DefaultWriteAheadLogDirectory  string = "/opt/universum/wal"
	DefaultWriteAheadLogBufferSize int64  = 1024 * 1024 // 1 MB
	DefaultWriteAheadLogFrequency  int64  = 5           // 5 seconds

	// Section:Logging
	DefaultServerLogFilePath string = "/var/log/universum/server.log"
	DefaultMinimumLogLevel   string = "INFO" // [DEBUG, INFO, WARN, ERROR, FATAL]

	// Section:Eviction
	DefaultAutoExpiryFrequency int64  = 2      // 2 seconds
	DefaultAutoEvictionPolicy  string = "NONE" // [NONE, LRU, LFU]

	// Section:Auth
	DefaultAuthenticationMode int64 = 1 // TBD
)
