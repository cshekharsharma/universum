package cluster

import "universum/config"

type Cluster struct {
	HeartbeatPort     uint16
	ReplicationFactor uint8
	Nodes             []*Node
	LeaderNode        uint64
}

func NewCluster() *Cluster {
	return &Cluster{
		HeartbeatPort:     uint16(config.Store.Cluster.HeartbeatPort),
		ReplicationFactor: uint8(config.Store.Cluster.ReplicationFactor),
		Nodes:             make([]*Node, 0),
	}
}

func InitCluster() {
	// join or form the cluster
}

func FormCluster() {
	// form a fresh cluster
}
func JoinCluster() {
	// join an exsiting cluster
}
