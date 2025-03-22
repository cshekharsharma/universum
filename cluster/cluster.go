package cluster

import (
	"sync"
	"time"
	"universum/config"
	"universum/internal/logger"
)

const (
	MaxHeartbeatRetryAttempts int = 3
)

var (
	clusterMu sync.Mutex
)

type RetryEntry struct {
	Node        *Node
	Attempts    int
	LastAttempt time.Time
}

type cluster struct {
	HeartbeatPort     uint16
	ReplicationFactor uint8
	Nodes             []*Node
	LeaderNodeId      uint64
	RetryQueue        map[uint64]*RetryEntry
	RetryInterval     time.Duration
}

var _cluster *cluster

func GetCluster() *cluster {
	clusterMu.Lock()
	defer clusterMu.Unlock()

	if _cluster == nil {
		_cluster = newCluster()
	}
	return _cluster
}

func newCluster() *cluster {
	return &cluster{
		HeartbeatPort:     uint16(config.Store.Cluster.HeartbeatPort),
		ReplicationFactor: uint8(config.Store.Cluster.ReplicationFactor),
		Nodes:             make([]*Node, 0),
		LeaderNodeId:      0,
		RetryQueue:        make(map[uint64]*RetryEntry),
		RetryInterval:     time.Duration(config.Store.Cluster.HeartbeatIntervalMs) * time.Millisecond,
	}
}

func (c *cluster) Init() {
	err := c.JoinCluster()

	if err == nil {
		heartbeatStopper = make(chan struct{})
		go StartGossipHeartbeat(heartbeatStopper)
		go c.startRetryQueueProcessor()
		return
	}

	// if joining cluster failed, form a new cluster
	err = c.FormCluster()
	if err != nil {
		logger.Get().Fatal("Failed to form cluster: %v", err)
	}
}

func (c *cluster) FormCluster() error {
	// form a fresh cluster
	return nil
}

func (c *cluster) JoinCluster() error {
	// join an exsiting cluster
	return nil
}

func (c *cluster) startRetryQueueProcessor() {
	defer func() {
		if r := recover(); r != nil {
			logger.Get().Error("Retry queue processor panicked: %v. Restarting...", r)
			c.startRetryQueueProcessor() // Respawn the processor
		}
	}()

	for {
		c.processRetryQueue()
		time.Sleep(c.RetryInterval) // Wait for the retry interval before processing again
	}
}

func (c *cluster) processRetryQueue() {
	for _, entry := range c.RetryQueue {
		if entry.Attempts >= MaxHeartbeatRetryAttempts {
			logger.Get().Info("Node %d exceeded max retry attempts, marking as unreachable", entry.Node.ID)
			continue
		}

		if time.Since(entry.LastAttempt) < c.RetryInterval {
			continue // Skip if retry interval has not passed
		}

		err := sendGossipMessage(entry.Node)
		entry.LastAttempt = time.Now()

		if err != nil {
			entry.Attempts++
			logger.Get().Warn("Retry failed for node %d: %v", entry.Node.ID, err)
		} else {
			delete(c.RetryQueue, entry.Node.ID) // Remove from retry queue if successful
			logger.Get().Info("Successfully reconnected to node %d", entry.Node.ID)
		}
	}
}
