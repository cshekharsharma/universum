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
	Server   *Server   `toml:"Server"`
	Cluster  *Cluster  `toml:"Cluster"`
	Storage  *Storage  `toml:"Storage"`
	Logging  *Logging  `toml:"Logging"`
	Eviction *Eviction `toml:"Eviction"`
	Auth     *Auth     `toml:"Auth"`
}

type Server struct {
	ServerPort              int64 `toml:"ServerPort"`
	MaxConnections          int64 `toml:"MaxConnections"`
	ConnectionWriteTimeout  int64 `toml:"ConnectionWriteTimeout"`
	RequestExecutionTimeout int64 `toml:"RequestExecutionTimeout"`
}

type Cluster struct {
	EnableCluster bool     `toml:"EnableCluster"`
	HeartbeatPort int64    `toml:"HeartbeatPort"`
	Hosts         []string `toml:"Hosts"`
}

type Memory struct {
	AllowedMemoryStorageLimit int64  `toml:"AllowedMemoryStorageLimit"`
	AutoSnapshotFrequency     int64  `toml:"AutoSnapshotFrequency"`
	SnapshotFileDirectory     string `toml:"SnapshotFileDirectory"`
	SnapshotCompressionAlgo   string `toml:"SnapshotCompressionAlgo"`
	RestoreSnapshotOnStart    bool   `toml:"RestoreSnapshotOnStart"`
}

type LSM struct {
	MemtableStorageType     string  `toml:"MemtableStorageType"`
	BloomFalsePositiveRate  float64 `toml:"BloomFalsePositiveRate"`
	BloomFilterMaxRecords   int64   `toml:"BloomFilterMaxRecords"`
	BlockCompressionAlgo    string  `toml:"BlockCompressionAlgo"`
	DataStorageDirectory    string  `toml:"DataStorageDirectory"`
	WriteBlockSize          int64   `toml:"WriteBlockSize"`
	WriteBufferSize         int64   `toml:"WriteBufferSize"`
	WriteAheadLogDirectory  string  `toml:"WriteAheadLogDirectory"`
	WriteAheadLogAsyncFlush bool    `toml:"WriteAheadLogAsyncFlush"`
	WriteAheadLogFrequency  int64   `toml:"WriteAheadLogFrequency"`
	WriteAheadLogBufferSize int64   `toml:"WriteAheadLogBufferSize"`
	BlockCacheMemoryLimit   int64   `toml:"BlockCacheMemoryLimit"`
}

type Storage struct {
	Memory *Memory `toml:"Memory"`
	LSM    *LSM    `toml:"LSM"`

	StorageEngine        string `toml:"StorageEngine"`
	MaxRecordSizeInBytes int64  `toml:"MaxRecordSizeInBytes"`
}

type Logging struct {
	LogFileDirectory string `toml:"LogFileDirectory"`
	MinimumLogLevel  string `toml:"MinimumLogLevel"`
}

type Eviction struct {
	AutoRecordExpiryFrequency int64  `toml:"AutoRecordExpiryFrequency"`
	AutoEvictionPolicy        string `toml:"RecordAutoEvictionPolicy"`
}

type Auth struct {
	AuthenticationEnabled bool   `toml:"AuthenticationEnabled"`
	DbUserName            string `toml:"DbUserName"`
	DbUserPassword        string `toml:"DbUserPassword"`
}
