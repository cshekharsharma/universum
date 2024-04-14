package engine

import "sync/atomic"

var ks_networkBytesSent int64 = 0
var ks_networkBytesReceived int64 = 0
var ks_commandsProcessed int64 = 0
var ks_totalKeyCount int64 = 0
var ks_keyCountWithTTL int64 = 0

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

func GetTotalKeyCount() int64 {
	return atomic.LoadInt64(&ks_totalKeyCount)
}

func AddTotalKeyCount(delta int64) {
	atomic.AddInt64(&ks_totalKeyCount, delta)
}

func ReduceTotalKeyCount(delta int64) {
	negativeDelta := ^delta + 1
	atomic.AddInt64(&ks_totalKeyCount, negativeDelta)
}

func GetKeyCountWithTTL() int64 {
	return atomic.LoadInt64(&ks_keyCountWithTTL)
}

func AddKeyCountWithTTL(delta int64) {
	atomic.AddInt64(&ks_keyCountWithTTL, delta)
}

func ReduceKeyCountWithTTL(delta int64) {
	negativeDelta := ^delta + 1
	atomic.AddInt64(&ks_keyCountWithTTL, negativeDelta)
}
