package engine

import (
	"fmt"
	"sync"
	"time"
	"universum/config"
	"universum/utils"
)

const (
	HealthyMemotyConsumptionRatio float64 = 1.01
)

var evictionMutex sync.Mutex

var evictionJobLastExecutedAt time.Time
var evictionJobExecutionFrequency time.Duration

type autoEvictionWorker struct {
	ExecutionErr error
}

func (w *autoEvictionWorker) startAutoEviction(evictionChan chan<- autoEvictionWorker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				w.ExecutionErr = err
			} else {
				w.ExecutionErr = fmt.Errorf("panic happened with %v", r)
			}
		} else {
			w.ExecutionErr = err
		}

		// emit message to the worker channel that worker is dying.
		evictionChan <- *w
	}()

	evictionJobLastExecutedAt = time.Now()

	for {
		nextScheduledTime := evictionJobLastExecutedAt.Add(evictionJobExecutionFrequency)

		if nextScheduledTime.Compare(time.Now()) < 1 {
			EvictRecords()
			evictionJobLastExecutedAt = time.Now()
		}

		time.Sleep(evictionJobExecutionFrequency)
	}
}

func EvictRecords() {
	evictionMutex.Lock()
	defer evictionMutex.Unlock()

	currMemUsage := utils.GetMemoryUsedByCurrentPID()
	allowedUsage := config.Store.Storage.Memory.AllowedMemoryStorageLimit

	policy := config.Store.Eviction.AutoEvictionPolicy
	if policy == config.EvictionPolicyNone || !isDbOverflown(currMemUsage, allowedUsage) {
		return
	}

	if policy == config.EvictionPolicyLRU {
		evictLRU(int64(currMemUsage), allowedUsage)
	}
}

func isDbOverflown(currUsage uint64, allowedUsage int64) bool {
	return float64(allowedUsage)*HealthyMemotyConsumptionRatio < float64(currUsage)
}

func evictLRU(currUsage int64, allowedUsage int64) {
	// @TODO: to be implemented
}
