package utils

import (
	"os"

	"github.com/shirou/gopsutil/v3/process"
)

const HEALTHY_MEMORY_CONSUMPTION_RATIO = 0.8

func GetMemoryUsedByCurrentPID() uint64 {
	pid := os.Getpid()
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return 0
	}

	memInfo, err := proc.MemoryInfo()
	if err != nil {
		return 0
	}

	return memInfo.RSS
}
