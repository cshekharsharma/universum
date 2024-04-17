package engine

import (
	"universum/config"
	"universum/utils"
)

const (
	EVICTION_POLICY_NONE string = "NONE"
	EVICTION_POLICY_LRU  string = "LRU"

	HEALTHY_MEMORY_CONSUMPTION_RATIO float64 = 0.99
)

func ShouldRunAutoEvictionNow() bool {
	policy := config.GetRecordAutoEvictionPolicy()

	if policy == EVICTION_POLICY_NONE {
		return false
	}

	currMemoryUsage := utils.GetMemoryUsedByCurrentPID()
	allowedUsage := config.GetAllowedMemoryStorageLimit()

	return float64(allowedUsage)*HEALTHY_MEMORY_CONSUMPTION_RATIO <= float64(currMemoryUsage)
}

func EvictRecords() {
	// @TODO: to be implemented
}
