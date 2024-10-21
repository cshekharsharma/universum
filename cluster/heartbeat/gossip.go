package heartbeat

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"time"
	"universum/cluster"
	"universum/resp3"
	"universum/server"
)

const DefaultNumGossipNodes int = 3
const DefaultRedundencyFactor int = 2

func StartGossipHeartbeat(nodes []*cluster.Node, intervalMs int64) {
	ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		gossipToRandomNodes(nodes)
	}
}

func gossipToRandomNodes(nodes []*cluster.Node) {
	clusterSize := len(nodes)
	gossipNodeCount := deriveGossipNodeCount(clusterSize)

	for i := 0; i < gossipNodeCount; i++ {
		randomNode := nodes[rand.Intn(clusterSize)]
		sendGossipMessage(randomNode)
	}
}

func sendGossipMessage(node *cluster.Node) {
	socket := fmt.Sprintf("%s.%d", node.Host, node.HeartbeatPort)
	addr, err := net.ResolveTCPAddr(server.NetworkTCP, socket)
	if err != nil {
		fmt.Println("Error resolving address for gossip:", err)
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

func deriveGossipNodeCount(clusterSize int) int {
	if clusterSize <= DefaultNumGossipNodes*DefaultNumGossipNodes {
		return DefaultNumGossipNodes
	}

	return int(math.Ceil(math.Sqrt(float64(clusterSize)) * float64(DefaultRedundencyFactor)))
}
