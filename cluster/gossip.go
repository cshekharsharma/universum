package cluster

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"time"
	"universum/config"
	"universum/internal/logger"
	"universum/resp3"
	"universum/server"
)

const DefaultNumGossipNodes int = 3
const DefaultRedundencyFactor int = 2

var heartbeatStopper chan struct{}

func StartGossipHeartbeat(done <-chan struct{}) {
	nodes := GetCluster().Nodes
	intervalMs := config.Store.Cluster.HeartbeatIntervalMs

	defer func(nodes []*Node, intervalMs int64, done <-chan struct{}) {
		if r := recover(); r != nil {
			logger.Get().Warn("StartGossipHeartbeat: Recovered from panic, will restart: %v", r)
			go StartGossipHeartbeat(done) // restart the heartbeat worker
		}
	}(nodes, intervalMs, done)

	ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gossipToRandomNodes(nodes)
		case <-done:
			logger.Get().Info("Stopping gossip heartbeat worker on request...")
			return
		}
	}
}

func gossipToRandomNodes(nodes []*Node) {
	clusterSize := len(nodes)
	gossipNodeCount := calculateGossipNodeCount(clusterSize)

	for i := 0; i < gossipNodeCount; i++ {
		randomNode := nodes[rand.Intn(clusterSize)]
		sendGossipMessage(randomNode)
	}
}

func sendGossipMessage(node *Node) error {
	socket := fmt.Sprintf("%s.%d", node.Host, node.HeartbeatPort)
	addr, err := net.ResolveTCPAddr(server.NetworkTCP, socket)
	if err != nil {
		logger.Get().Error("Error resolving gossip address [%s]: %v", socket, err)
		return err
	}

	conn, err := net.DialTCP(server.NetworkTCP, nil, addr)
	if err != nil {
		logger.Get().Error("Error dialing UDP for gossip:", err)
		return err
	}
	defer conn.Close()

	heartbeatMsg := NewHeartbeat(node.ID, node.Status, node.StartTime)
	message, _ := resp3.Encode(heartbeatMsg)

	_, err = conn.Write([]byte(message))
	if err != nil {
		logger.Get().Error("Error sending gossip heartbeat:", err)
	}

	logger.Get().Info("Sent gossip heartbeat to", node.Host)
	return nil
}

func calculateGossipNodeCount(clusterSize int) int {
	if clusterSize <= DefaultNumGossipNodes*DefaultNumGossipNodes {
		return DefaultNumGossipNodes
	}

	// gossip to sqrt(N) * R nodes
	return int(math.Ceil(math.Sqrt(float64(clusterSize)) * float64(DefaultRedundencyFactor)))
}
