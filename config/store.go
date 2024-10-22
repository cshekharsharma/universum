package config

// InvalidNumericValue is a constant representing an invalid numeric value
// returned when a requested configuration key is not found.
const InvalidNumericValue = -99999999

// InvalidEpochValue is a constant representing an invalid epoch value,
// used when a date or time-related key is not found or improperly formatted.
const InvalidEpochValue = 0

// InfiniteExpiryTime is a constant representing an infinite expiry time
// for a record in the database.
const InfiniteExpiryTime int64 = 4102444800

type Config struct {
	Server   *Server   `toml:"Server"`   // Configuration for server-related settings
	Cluster  *Cluster  `toml:"Cluster"`  // Cluster configuration settings
	Storage  *Storage  `toml:"Storage"`  // Storage engine settings
	Logging  *Logging  `toml:"Logging"`  // Log file and log level settings
	Eviction *Eviction `toml:"Eviction"` // Automatic eviction and expiry policy settings
	Auth     *Auth     `toml:"Auth"`     // Authentication-related settings
}

type Server struct {
	ServerPort              int64  `toml:"ServerPort"`              // Port where the server listens for incoming connections
	MaxConnections          int64  `toml:"MaxConnections"`          // Maximum number of concurrent client connections allowed
	ConnectionWriteTimeout  int64  `toml:"ConnectionWriteTimeout"`  // Maximum duration (seconds) to wait before timing out a write operation
	RequestExecutionTimeout int64  `toml:"RequestExecutionTimeout"` // Maximum duration (seconds) to wait before timing out a request execution
	EnableTLS               bool   `toml:"EnableTLS"`               // Enable or disable TLS for secure communication
	TLSCertFilePath         string `toml:"TLSCertFilePath"`         // Path to the TLS certificate file
	TLSKeyFilePath          string `toml:"TLSKeyFilePath"`          // Path to the TLS key file
}

type Cluster struct {
	EnableCluster     bool     `toml:"EnableCluster"`     // Enable or disable cluster functionality
	Hosts             []string `toml:"Hosts"`             // List of seed or known nodes in the cluster
	HeartbeatMode     string   `toml:"HeartbeatMode"`     // Mode of heartbeats; can be "gossip" or "multicast"
	MulticastAddress  string   `toml:"MulticastAddress"`  // IP address used for multicast communication (if enabled)
	MulticastPort     int64    `toml:"MulticastPort"`     // Port number used for multicast communication
	HeartbeatPort     int64    `toml:"HeartbeatPort"`     // Port number used for sending heartbeat signals
	GossipIntervalMs  int64    `toml:"GossipIntervalMs"`  // Time interval in milliseconds between gossip messages
	ReplicationFactor int64    `toml:"ReplicationFactor"` // Number of copies (replicas) of each record across the cluster
}

type Memory struct {
	AllowedMemoryStorageLimit int64  `toml:"AllowedMemoryStorageLimit"` // Maximum memory allocation for in-memory storage
	AutoSnapshotFrequency     int64  `toml:"AutoSnapshotFrequency"`     // Frequency in seconds to auto-generate snapshots of memory state
	SnapshotFileDirectory     string `toml:"SnapshotFileDirectory"`     // Directory to store memory snapshots
	SnapshotCompressionAlgo   string `toml:"SnapshotCompressionAlgo"`   // Compression algorithm used for snapshots
	RestoreSnapshotOnStart    bool   `toml:"RestoreSnapshotOnStart"`    // Whether to restore a snapshot at server startup
}

type LSM struct {
	MemtableStorageType     string  `toml:"MemtableStorageType"`     // Type of storage structure used for the MemTable
	BloomFalsePositiveRate  float64 `toml:"BloomFalsePositiveRate"`  // Probability of false positives for Bloom filters
	BloomFilterMaxRecords   int64   `toml:"BloomFilterMaxRecords"`   // Maximum number of records tracked by a Bloom filter
	BlockCompressionAlgo    string  `toml:"BlockCompressionAlgo"`    // Compression algorithm used for SSTable blocks
	DataStorageDirectory    string  `toml:"DataStorageDirectory"`    // Directory path for storing data files/sstables
	WriteBlockSize          int64   `toml:"WriteBlockSize"`          // Size of each block written to SSTables
	WriteBufferSize         int64   `toml:"WriteBufferSize"`         // Size of the buffer for writing data (memtable size)
	WriteAheadLogDirectory  string  `toml:"WriteAheadLogDirectory"`  // Directory for write-ahead logs (WAL)
	WriteAheadLogAsyncFlush bool    `toml:"WriteAheadLogAsyncFlush"` // Enable/disable asynchronous flushing of WAL
	WriteAheadLogFrequency  int64   `toml:"WriteAheadLogFrequency"`  // Frequency of WAL flushes
	WriteAheadLogBufferSize int64   `toml:"WriteAheadLogBufferSize"` // Buffer size for write-ahead logs
	BlockCacheMemoryLimit   int64   `toml:"BlockCacheMemoryLimit"`   // Maximum memory allowed for block cache
}

type Storage struct {
	Memory *Memory `toml:"Memory"` // Configuration settings for in-memory storage
	LSM    *LSM    `toml:"LSM"`    // Configuration settings for LSM tree storage

	StorageEngine        string `toml:"StorageEngine"`        // Name of the storage engine to be used (e.g., LSM or Memory)
	MaxRecordSizeInBytes int64  `toml:"MaxRecordSizeInBytes"` // Maximum allowed size for each record
}

type Logging struct {
	LogFileDirectory string `toml:"LogFileDirectory"` // Directory where log files are stored
	MinimumLogLevel  string `toml:"MinimumLogLevel"`  // Minimum level of log messages to be recorded (e.g., DEBUG, INFO, ERROR)
}

type Eviction struct {
	AutoRecordExpiryFrequency int64  `toml:"AutoRecordExpiryFrequency"` // Frequency of automatic record expiry checks
	AutoEvictionPolicy        string `toml:"RecordAutoEvictionPolicy"`  // Policy for auto-eviction of records (e.g., LRU, FIFO)
}

type Auth struct {
	AuthenticationEnabled bool   `toml:"AuthenticationEnabled"` // Enable or disable authentication mechanism
	DbUserName            string `toml:"DbUserName"`            // Username for database access
	DbUserPassword        string `toml:"DbUserPassword"`        // Password for database access
}
