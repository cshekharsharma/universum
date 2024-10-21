package cluster

import (
	"net"
	"time"
	"universum/config"
	"universum/crypto"
	"universum/utils"
)

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

type Node struct {
	ID            uint64
	Host          string
	Status        uint8
	IsLeader      bool
	HeartbeatPort uint16
	Partitions    []int64

	LastPingedAt  int64
	LastUpdatedAt int64
	StartTime     int64

	FailureCount uint8
	LastFailedAt int64
}

func NewClusterNode() (*Node, error) {
	iface, addr, err := utils.GetPrimaryNetworkInterface()
	if err != nil {
		return nil, err
	}

	node := &Node{
		ID:            GenerateUniqueNodeID(iface, addr),
		Host:          addr.String(),
		Status:        NodeStatusInited,
		IsLeader:      true, // By default, the first node is the leader
		HeartbeatPort: uint16(config.Store.Cluster.HeartbeatPort),
		Partitions:    make([]int64, 0),
		LastPingedAt:  time.Now().Unix(),
		LastUpdatedAt: time.Now().Unix(),
		StartTime:     time.Now().Unix(),
	}

	return node, nil
}

func GenerateUniqueNodeID(iface *net.Interface, addr net.Addr) uint64 {
	ipStr := addr.String()
	mac := iface.HardwareAddr.String()

	return crypto.MurmurHash64([]byte(ipStr+mac), digestSeed)
}
