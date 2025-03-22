package entity

import "encoding/json"

type InfoStats struct {
	StatsGenerationTime int64
	Server              *ServerStats
	Clients             *ClientStats
	Persistence         *PersistenceStats
	CpuInfo             *CpuStats
	Network             *NetworkStats
	Keyspace            *KeyspaceStats
}

func (is *InfoStats) ToString() string {
	if jsonstr, err := json.Marshal(is); err == nil {
		return string(jsonstr)
	}
	return ""
}

type ServerStats struct {
	BuildVersion string
	TCPPort      int64
	ConfigFile   string
	OSName       string
	ArchBits     string
	ClockTime    string
	TimeZone     string
	StartedAt    string
	ServerState  string
}

type ClientStats struct {
	MaxAllowedConnections    int64
	MaxConnectionConcurrency int64
	ConnectedClients         int64
}

type PersistenceStats struct {
	AutoSnapshotFrequency     string
	SnapshotFileDirectory     string
	SnapshotSizeInBytes       int64
	LastSnapshotTakenAt       string
	LastSnapshotLatency       string
	LastSnapshotReplayedAt    string
	LastSnapshotReplayLatency string
	TotalKeysReplayed         int64
}

type CpuStats struct {
	CpuCount                  int64
	CpuLoadPercent            float64
	TotalMemory               uint64
	FreeMemory                uint64
	AllowedMemoryStorageLimit uint64
	MemoryStorageConsumption  uint64
}

type NetworkStats struct {
	CommandsProcessed    int64
	NetworkBytesSent     int64
	NetworkBytesReceived int64
}

type KeyspaceStats struct {
	TotalKeyCount   int64
	KeyCountWithTTL int64
}
