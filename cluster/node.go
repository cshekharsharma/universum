package cluster

import "time"

const (
	NodeStatusInited      uint8 = 0
	NodeStatusJoining     uint8 = 1
	NodeStatusJoined      uint8 = 2
	NodeStatusActive      uint8 = 3
	NodeStatusInactive    uint8 = 4
	NodeStatusUnreachable uint8 = 5
	NodeStatusLeaving     uint8 = 6
	NodeStatusLeft        uint8 = 7
)

type ClusterNode struct {
	ID           string
	Host         string
	Paritions    []int64
	Status       uint8
	LastPingedAt time.Time
}

func NewClusterNode(id, host string) *ClusterNode {
	return &ClusterNode{
		ID:           id,
		Host:         host,
		Paritions:    make([]int64, 0),
		Status:       NodeStatusInited,
		LastPingedAt: time.Now(),
	}
}
