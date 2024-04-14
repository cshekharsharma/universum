package engine

import (
	"go/build"
	"math"
	"time"
	"universum/config"
	"universum/consts"
	"universum/engine/entity"
	"universum/utils"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var DatabaseInfoStats *entity.InfoStats

func InitInfoStatistics() {
	timezone, _ := time.Now().Zone()

	DatabaseInfoStats = &entity.InfoStats{
		StatsGenerationTime: utils.GetCurrentEPochTime(),
		Server: &entity.ServerStats{
			BuildVersion: consts.SERVER_VERSION,
			TCPPort:      config.GetServerPort(),
			ClockTime:    utils.GetCurrentReadableTime(),
			ConfigFile:   config.DEFAULT_CONFIG_NAME,
			OSName:       build.Default.GOOS,
			ArchBits:     build.Default.GOARCH,
			TimeZone:     timezone,
			ServerState:  consts.GetServerStateAsString(),
			StartedAt:    utils.GetCurrentReadableTime(),
		},

		Clients: &entity.ClientStats{
			MaxAllowedConnections:    config.GetMaxClientConnections(),
			MaxConnectionConcurrency: config.GetServerConcurrencyLimit(config.GetMaxClientConnections()),
			ConnectedClients:         -1,
		},

		Persistence: &entity.PersistenceStats{
			AutoSnapshotFrequency: config.GetAutoSnapshotFrequency().String(),
			SnapshotFilePath:      config.GetTransactionLogFilePath(),
			LastSnapshotTakenAt:   snapshotJobLastExecutedAt.String(),
		},

		CpuInfo: &entity.CpuStats{
			CpuCount:       0,
			CpuLoadPercent: 0,
			TotalMemory:    0,
			UsedMemory:     0,
		},

		Network: &entity.NetworkStats{
			CommandsProcessed:    GetCommandsProcessed(),
			NetworkBytesSent:     GetNetworkBytesSent(),
			NetworkBytesReceived: GetNetworkBytesReceived(),
		},

		Keyspace: &entity.KeyspaceStats{
			TotalKeyCount:   GetTotalKeyCount(),
			KeyCountWithTTL: GetKeyCountWithTTL(),
		},
	}
}

func GetDatabaseInfoStatistics() *entity.InfoStats {
	DatabaseInfoStats.Server.ClockTime = utils.GetCurrentReadableTime()

	if cpucount, err := cpu.Counts(true); err == nil {
		DatabaseInfoStats.CpuInfo.CpuCount = int64(cpucount)
	}

	if percentages, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(percentages) > 0 {
		DatabaseInfoStats.CpuInfo.CpuLoadPercent = math.Round(percentages[0])
	}

	if virtualMemory, err := mem.VirtualMemory(); err == nil {
		DatabaseInfoStats.CpuInfo.TotalMemory = virtualMemory.Total
		DatabaseInfoStats.CpuInfo.UsedMemory = virtualMemory.Used
	}

	DatabaseInfoStats.Network.CommandsProcessed = GetCommandsProcessed()
	DatabaseInfoStats.Network.NetworkBytesSent = GetNetworkBytesSent()
	DatabaseInfoStats.Network.NetworkBytesReceived = GetNetworkBytesReceived()

	return DatabaseInfoStats
}
