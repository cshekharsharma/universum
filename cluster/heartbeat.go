package cluster

import (
	"errors"
	"time"
)

// Heartbeat represents the structure for storing heartbeat information.
type Heartbeat struct {
	id  uint64 // Unique identifier for the heartbeat
	st  uint8  // Status of the heartbeat
	ts  int64  // Timestamp of the heartbeat
	ut  int64  // Uptime of the heartbeat
	gen uint64 // Generation number of the heartbeat
	pmv uint64 // Partition map version of the heartbeat
}

func NewHeartbeat(id uint64, status uint8, uptime int64) *Heartbeat {
	return &Heartbeat{
		id:  id,
		st:  status,
		ts:  time.Now().Unix(),
		ut:  uptime,
		gen: 0,
		pmv: 0,
	}
}

func (h *Heartbeat) GetNodeID() uint64 {
	return h.id
}

func (h *Heartbeat) GetStatus() uint8 {
	return h.st
}

func (h *Heartbeat) GetRequestTimestamp() int64 {
	return h.ts
}

func (h *Heartbeat) GetNodeUptime() int64 {
	return h.ut
}

func (h *Heartbeat) GetClusterGeneration() uint64 {
	return h.gen
}

func (h *Heartbeat) GetPreviousMajorVersion() uint64 {
	return h.pmv
}

func (h *Heartbeat) SetNodeID(id uint64) {
	h.id = id
}

func (h *Heartbeat) SetStatus(st uint8) {
	h.st = st
}

func (h *Heartbeat) SetRequestTimestamp(ts int64) {
	h.ts = ts
}

func (h *Heartbeat) SetNodeUptime(ut int64) {
	h.ut = ut
}

func (h *Heartbeat) SetClusterGeneration(gen uint64) {
	h.gen = gen
}

func (h *Heartbeat) SetPreviousMajorVersion(pmv uint64) {
	h.pmv = pmv
}

func (h *Heartbeat) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":  h.id,
		"st":  h.st,
		"ts":  h.ts,
		"ut":  h.ut,
		"gen": h.gen,
		"pmv": h.pmv,
	}
}

func HeartbeatEntityFromMap(input map[string]interface{}) (*Heartbeat, error) {
	if input == nil {
		return nil, errors.New("input map is nil")
	}

	heartbeat := &Heartbeat{}

	if id, ok := input["id"].(uint64); ok {
		heartbeat.id = id
	}

	if st, ok := input["st"].(uint8); ok {
		heartbeat.st = st
	}

	if ts, ok := input["ts"].(int64); ok {
		heartbeat.ts = ts
	}

	if ut, ok := input["ut"].(int64); ok {
		heartbeat.ut = ut
	}

	if gen, ok := input["gen"].(uint64); ok {
		heartbeat.gen = gen
	}

	if pmv, ok := input["pmv"].(uint64); ok {
		heartbeat.pmv = pmv
	}

	return heartbeat, nil
}
