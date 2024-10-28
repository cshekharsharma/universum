package heartbeat

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"time"
	"universum/cluster"
	"universum/internal/logger"
	"universum/resp3"
	"universum/server"
)

const DefaultNumGossipNodes int = 3
const DefaultRedundencyFactor int = 2

var heartbeatStopper = make(chan struct{})

func StartGossipHeartbeat(nodes []*cluster.Node, intervalMs int64, done <-chan struct{}) {
	defer func(nodes []*cluster.Node, intervalMs int64, done <-chan struct{}) {
		if r := recover(); r != nil {
			logger.Get().Warn("StartGossipHeartbeat: Recovered from panic, will restart: %v", r)
			go StartGossipHeartbeat(nodes, intervalMs, done) // restart the heartbeat worker
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

func gossipToRandomNodes(nodes []*cluster.Node) {
	clusterSize := len(nodes)
	gossipNodeCount := calculateGossipNodeCount(clusterSize)

	for i := 0; i < gossipNodeCount; i++ {
		randomNode := nodes[rand.Intn(clusterSize)]
		sendGossipMessage(randomNode)
	}
}

func sendGossipMessage(node *cluster.Node) {
	socket := fmt.Sprintf("%s.%d", node.Host, node.HeartbeatPort)
	addr, err := net.ResolveTCPAddr(server.NetworkTCP, socket)
	if err != nil {
		logger.Get().Fatal("Error resolving gossip address [%s]: %v", socket, err)
		return
	}

	conn, err := net.DialTCP(server.NetworkTCP, nil, addr)
	if err != nil {
		fmt.Println("Error dialing UDP for gossip:", err)
		return
	}
	defer conn.Close()

	heartbeatMsg := NewHeartbeat(node.ID, node.Status, node.StartTime)
	message, _ := resp3.Encode(heartbeatMsg)

	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending gossip heartbeat:", err)
	}
	fmt.Println("Sent gossip heartbeat to", node.Host)
}

func calculateGossipNodeCount(clusterSize int) int {
	if clusterSize <= DefaultNumGossipNodes*DefaultNumGossipNodes {
		return DefaultNumGossipNodes
	}

	// gossip to sqrt(N) * R nodes
	return int(math.Ceil(math.Sqrt(float64(clusterSize)) * float64(DefaultRedundencyFactor)))
}
