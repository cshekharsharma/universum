package engine

import (
	"go/build"
	"math"
	"sync/atomic"
	"time"
	"universum/config"
	"universum/entity"
	"universum/utils"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var DatabaseInfoStats *entity.InfoStats

var ks_networkBytesSent int64 = 0
var ks_networkBytesReceived int64 = 0
var ks_commandsProcessed int64 = 0

func GetNetworkBytesSent() int64 {
	return atomic.LoadInt64(&ks_networkBytesSent)
}

func AddNetworkBytesSent(delta int64) {
	atomic.AddInt64(&ks_networkBytesSent, delta)
}

func GetNetworkBytesReceived() int64 {
	return atomic.LoadInt64(&ks_networkBytesReceived)
}

func AddNetworkBytesReceived(delta int64) {
	atomic.AddInt64(&ks_networkBytesReceived, delta)
}

func GetCommandsProcessed() int64 {
	return atomic.LoadInt64(&ks_commandsProcessed)
}

func AddCommandsProcessed(delta int64) {
	atomic.AddInt64(&ks_commandsProcessed, delta)
}

func InitInfoStatistics() {
	timezone, _ := time.Now().Zone()

	DatabaseInfoStats = &entity.InfoStats{
		StatsGenerationTime: utils.GetCurrentEPochTime(),
		Server: &entity.ServerStats{
			BuildVersion: entity.SERVER_VERSION,
			TCPPort:      config.Store.Server.ServerPort,
			ClockTime:    utils.GetCurrentReadableTime(),
			ConfigFile:   config.DefaultConfigFilePath,
			OSName:       build.Default.GOOS,
			ArchBits:     build.Default.GOARCH,
			TimeZone:     timezone,
			ServerState:  entity.GetServerStateAsString(),
			StartedAt:    utils.GetCurrentReadableTime(),
		},

		Clients: &entity.ClientStats{
			MaxAllowedConnections: int64(config.Store.Server.MaxConnections),
			ConnectedClients:      0,
		},

		Persistence: &entity.PersistenceStats{
			AutoSnapshotFrequency: time.Duration(config.Store.Storage.Memory.AutoSnapshotFrequency).String(),
			SnapshotFileDirectory: config.Store.Storage.Memory.SnapshotFileDirectory,
			LastSnapshotTakenAt:   snapshotJobLastExecutedAt.String(),
		},

		CpuInfo: &entity.CpuStats{},

		Network: &entity.NetworkStats{
			CommandsProcessed:    GetCommandsProcessed(),
			NetworkBytesSent:     GetNetworkBytesSent(),
			NetworkBytesReceived: GetNetworkBytesReceived(),
		},

		Keyspace: &entity.KeyspaceStats{
			TotalKeyCount:   0,
			KeyCountWithTTL: 0,
		},
	}
}

func GetDatabaseInfoStatistics() *entity.InfoStats {
	DatabaseInfoStats.Server.ConfigFile = config.ConfigFilePath
	DatabaseInfoStats.Server.ServerState = entity.GetServerStateAsString()
	DatabaseInfoStats.Server.ClockTime = utils.GetCurrentReadableTime()

	if cpucount, err := cpu.Counts(true); err == nil {
		DatabaseInfoStats.CpuInfo.CpuCount = int64(cpucount)
	}

	if percentages, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(percentages) > 0 {
		DatabaseInfoStats.CpuInfo.CpuLoadPercent = math.Round(percentages[0])
	}

	if virtualMemory, err := mem.VirtualMemory(); err == nil {
		DatabaseInfoStats.CpuInfo.TotalMemory = virtualMemory.Total
		DatabaseInfoStats.CpuInfo.FreeMemory = virtualMemory.Total - virtualMemory.Used
		DatabaseInfoStats.CpuInfo.AllowedMemoryStorageLimit = uint64(config.Store.Storage.Memory.AllowedMemoryStorageLimit)
		DatabaseInfoStats.CpuInfo.MemoryStorageConsumption = utils.GetMemoryUsedByCurrentPID()
	}

	DatabaseInfoStats.Network.CommandsProcessed = GetCommandsProcessed()
	DatabaseInfoStats.Network.NetworkBytesSent = GetNetworkBytesSent()
	DatabaseInfoStats.Network.NetworkBytesReceived = GetNetworkBytesReceived()

	activeConnections := entity.GetActiveTCPConnectionCount()
	DatabaseInfoStats.Clients.ConnectedClients = activeConnections

	if activeConnections < 0 {
		DatabaseInfoStats.Clients.ConnectedClients = 0
	}

	return DatabaseInfoStats
}
