package consts

import "sync/atomic"

var activeTCPConnections int64 = 0

func GetActiveTCPConnectionCount() int64 {
	return atomic.LoadInt64(&activeTCPConnections)
}

func IncrementActiveTCPConnection() {
	atomic.AddInt64(&activeTCPConnections, 1)
}

func DecrementActiveTCPConnection() {
	atomic.AddInt64(&activeTCPConnections, -1)
}
