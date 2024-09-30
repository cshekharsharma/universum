package config

// InvalidNumericValue is a constant representing an invalid numeric value
// returned when a requested configuration key is not found.
const InvalidNumericValue = -99999999

// InvalidEpochValue is a constant representing an invalid epoch value,
// used when a date or time-related key is not found or improperly formatted.
const InvalidEpochValue = 0

type Config struct {
	Server   *Server   `toml:"Server"`
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

type Memory struct {
	AllowedMemoryStorageLimit int64  `toml:"AllowedMemoryStorageLimit"`
	AutoSnapshotFrequency     int64  `toml:"AutoSnapshotFrequency"`
	SnapshotFileDirectory     string `toml:"SnapshotFileDirectory"`
	SnapshotCompressionAlgo   string `toml:"SnapshotCompressionAlgo"`
	RestoreSnapshotOnStart    bool   `toml:"RestoreSnapshotOnStart"`
}

type LSM struct {
	MemtableStorageType     string `toml:"MemtableStorageType"`
	WriteBlockSize          int64  `toml:"WriteBlockSize"`
	DataStorageDirectory    string `toml:"DataStorageDirectory"`
	WriteAheadLogDirectory  string `toml:"WriteAheadLogDirectory"`
	WriteAheadLogFrequency  int64  `toml:"WriteAheadLogFrequency"`
	WriteAheadLogBufferSize int64  `toml:"WriteAheadLogBufferSize"`
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
