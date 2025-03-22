package cluster

import (
	"testing"
	"time"
)

func TestNewHeartbeat(t *testing.T) {
	id := uint64(12345)
	status := uint8(1)
	uptime := int64(3600)

	heartbeat := NewHeartbeat(id, status, uptime)

	if heartbeat.GetNodeID() != id {
		t.Errorf("Expected NodeID %d, got %d", id, heartbeat.GetNodeID())
	}

	if heartbeat.GetStatus() != status {
		t.Errorf("Expected Status %d, got %d", status, heartbeat.GetStatus())
	}

	if heartbeat.GetNodeUptime() != uptime {
		t.Errorf("Expected Uptime %d, got %d", uptime, heartbeat.GetNodeUptime())
	}

	if diff := time.Now().Unix() - heartbeat.GetRequestTimestamp(); diff > 1 {
		t.Errorf("Expected current timestamp, got a difference of %d seconds", diff)
	}

	if heartbeat.GetClusterGeneration() != 0 {
		t.Errorf("Expected ClusterGeneration 0, got %d", heartbeat.GetClusterGeneration())
	}

	if heartbeat.GetPreviousMajorVersion() != 0 {
		t.Errorf("Expected PreviousMajorVersion 0, got %d", heartbeat.GetPreviousMajorVersion())
	}
}

func TestHeartbeatGettersAndSetters(t *testing.T) {
	heartbeat := NewHeartbeat(12345, 1, 3600)

	heartbeat.SetNodeID(54321)
	if heartbeat.GetNodeID() != 54321 {
		t.Errorf("Expected NodeID 54321, got %d", heartbeat.GetNodeID())
	}

	heartbeat.SetStatus(2)
	if heartbeat.GetStatus() != 2 {
		t.Errorf("Expected Status 2, got %d", heartbeat.GetStatus())
	}

	currentTime := time.Now().Unix()
	heartbeat.SetRequestTimestamp(currentTime)
	if heartbeat.GetRequestTimestamp() != currentTime {
		t.Errorf("Expected RequestTimestamp %d, got %d", currentTime, heartbeat.GetRequestTimestamp())
	}

	heartbeat.SetNodeUptime(7200)
	if heartbeat.GetNodeUptime() != 7200 {
		t.Errorf("Expected NodeUptime 7200, got %d", heartbeat.GetNodeUptime())
	}

	heartbeat.SetClusterGeneration(1001)
	if heartbeat.GetClusterGeneration() != 1001 {
		t.Errorf("Expected ClusterGeneration 1001, got %d", heartbeat.GetClusterGeneration())
	}

	heartbeat.SetPreviousMajorVersion(42)
	if heartbeat.GetPreviousMajorVersion() != 42 {
		t.Errorf("Expected PreviousMajorVersion 42, got %d", heartbeat.GetPreviousMajorVersion())
	}
}

func TestHeartbeatToMap(t *testing.T) {
	heartbeat := NewHeartbeat(12345, 1, 3600)
	heartbeat.SetClusterGeneration(1001)
	heartbeat.SetPreviousMajorVersion(42)

	hbMap := heartbeat.ToMap()

	if hbMap["id"] != uint64(12345) {
		t.Errorf("Expected id 12345, got %v", hbMap["id"])
	}

	if hbMap["st"] != uint8(1) {
		t.Errorf("Expected status 1, got %v", hbMap["st"])
	}

	if hbMap["ts"] != heartbeat.GetRequestTimestamp() {
		t.Errorf("Expected timestamp %d, got %v", heartbeat.GetRequestTimestamp(), hbMap["ts"])
	}

	if hbMap["ut"] != int64(3600) {
		t.Errorf("Expected uptime 3600, got %v", hbMap["ut"])
	}

	if hbMap["gen"] != uint64(1001) {
		t.Errorf("Expected cluster generation 1001, got %v", hbMap["gen"])
	}

	if hbMap["pmv"] != uint64(42) {
		t.Errorf("Expected previous major version 42, got %v", hbMap["pmv"])
	}
}

func TestHeartbeatEntityFromMap(t *testing.T) {
	hbMap := map[string]interface{}{
		"id":  uint64(12345),
		"st":  uint8(1),
		"ts":  time.Now().Unix(),
		"ut":  int64(3600),
		"gen": uint64(1001),
		"pmv": uint64(42),
	}

	heartbeat, err := HeartbeatEntityFromMap(hbMap)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if heartbeat.GetNodeID() != uint64(12345) {
		t.Errorf("Expected NodeID 12345, got %d", heartbeat.GetNodeID())
	}

	if heartbeat.GetStatus() != uint8(1) {
		t.Errorf("Expected Status 1, got %d", heartbeat.GetStatus())
	}

	if heartbeat.GetRequestTimestamp() != hbMap["ts"].(int64) {
		t.Errorf("Expected timestamp %v, got %d", hbMap["ts"], heartbeat.GetRequestTimestamp())
	}

	if heartbeat.GetNodeUptime() != int64(3600) {
		t.Errorf("Expected Uptime 3600, got %d", heartbeat.GetNodeUptime())
	}

	if heartbeat.GetClusterGeneration() != uint64(1001) {
		t.Errorf("Expected ClusterGeneration 1001, got %d", heartbeat.GetClusterGeneration())
	}

	if heartbeat.GetPreviousMajorVersion() != uint64(42) {
		t.Errorf("Expected PreviousMajorVersion 42, got %d", heartbeat.GetPreviousMajorVersion())
	}
}

func TestHeartbeatEntityFromMapNilInput(t *testing.T) {
	_, err := HeartbeatEntityFromMap(nil)
	if err == nil {
		t.Errorf("Expected error when input map is nil, got nil")
	}
}

func TestHeartbeatEntityFromMapPartialInput(t *testing.T) {
	hbMap := map[string]interface{}{
		"id": uint64(12345),
		"ut": int64(3600),
	}

	heartbeat, err := HeartbeatEntityFromMap(hbMap)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if heartbeat.GetNodeID() != uint64(12345) {
		t.Errorf("Expected NodeID 12345, got %d", heartbeat.GetNodeID())
	}

	if heartbeat.GetNodeUptime() != int64(3600) {
		t.Errorf("Expected Uptime 3600, got %d", heartbeat.GetNodeUptime())
	}

	if heartbeat.GetStatus() != uint8(0) {
		t.Errorf("Expected default Status 0, got %d", heartbeat.GetStatus())
	}

	if heartbeat.GetRequestTimestamp() != int64(0) {
		t.Errorf("Expected default RequestTimestamp 0, got %d", heartbeat.GetRequestTimestamp())
	}

	if heartbeat.GetClusterGeneration() != uint64(0) {
		t.Errorf("Expected default ClusterGeneration 0, got %d", heartbeat.GetClusterGeneration())
	}

	if heartbeat.GetPreviousMajorVersion() != uint64(0) {
		t.Errorf("Expected default PreviousMajorVersion 0, got %d", heartbeat.GetPreviousMajorVersion())
	}
}
