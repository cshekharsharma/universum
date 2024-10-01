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

	DefaultConfigFileDirectory string = "/etc/universum"
	DefaultConfigFileName      string = "config.ini"
	DefaultConfigFilePath      string = "/etc/universum/config.toml"
	DefaultWALFileName         string = "writeahead.aof"
	DefaultServerLogFile       string = "universum.log"

	// Constants for valid field values
	StorageEngineMemory   string = "MEMORY"
	StorageEngineLSM      string = "LSM"
	MemtableStorageTypeLB string = "LB"   // skip list + bloom filter
	MemtableStorageTypeLM string = "LM"   // sorted list + sync.Map
	CompressionAlgoNone   string = "NONE" // no compression
	CompressionAlgoLZ4    string = "LZ4"  // LZ4 compression

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
	DefultMaxClientConnections    int64 = 10000
	DefaultConnectionWriteTimeout int64 = 10 // 10 seconds
	DefaultRequestExecTimeout     int64 = 10 // 10 seconds

	// Storage
	DefaultStorageEngine        string = "MEMORY"    // [MEMORY, LSM]
	DefaultMaxRecordSizeInBytes int64  = 1024 * 1024 // 1 MB

	// Storage.Memory
	AllowedMemoryStorageLimit      int64  = 1024 * 1024 * 1024 // 1 GB
	DefaultAutoSnapshotFrequency   int64  = 10                 // 10 seconds
	DefaultSnapshotFileDirectory   string = "/opt/universum/snapshot"
	DefaultRestoreSnapshotOnStart  bool   = true
	DefaultSnapshotCompressionAlgo string = "LZ4"

	// Storage.LSM
	DefaultMemtableStorageType     string  = MemtableStorageTypeLB
	DefaultMaxMemtableRecords      int64   = 1000000 // 1 million
	DefaultBloomFalsePositiveRate  float64 = 0.01    // 1%
	DefaultWriteBlockSize          int64   = 65536   // 64 KB
	DefaultBlockCompressionAlgo    string  = CompressionAlgoLZ4
	DefaultDataStorageDirectory    string  = "/opt/universum/data"
	DefaultWriteAheadLogDirectory  string  = "/opt/universum/wal"
	DefaultWriteAheadLogBufferSize int64   = 1024 * 1024 // 1 MB
	DefaultWriteAheadLogFrequency  int64   = 5           // 5 seconds

	// Section:Logging
	LogLevelDebug string = "DEBUG"
	LogLevelInfo  string = "INFO"
	LogLevelWarn  string = "WARN"
	LogLevelError string = "ERROR"
	LogLevelFatal string = "FATAL"

	DefaultLogFileDirectory string = "/var/log/universum"
	DefaultMinimumLogLevel  string = LogLevelInfo

	// Section:Eviction
	EvictionPolicyLRU  = "LRU"
	EvictionPolicyNone = "NONE"

	DefaultRecordAutoExpiryFrequency int64  = 2 // 2 seconds
	DefaultAutoEvictionPolicy        string = EvictionPolicyNone

	// Section:Auth
	DefaultAuthenticationMode int64 = 1 // TBD
)

var AllowedLogLevels []string = []string{
	LogLevelDebug,
	LogLevelInfo,
	LogLevelWarn,
	LogLevelError,
	LogLevelFatal,
}

var AllowedAutoEvictionPolicy []string = []string{
	EvictionPolicyLRU,
	EvictionPolicyNone,
}

var AllowedCompressionAlgos []string = []string{
	CompressionAlgoNone,
	CompressionAlgoLZ4,
}
