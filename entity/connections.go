package entity

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Connection struct {
	Conn       *net.TCPConn
	RemoteAddr string
	CreatedAt  int64
}

var activeTCPConnections int64 = 0
var activeConnections = sync.Map{}

func GetActiveTCPConnectionCount() int64 {
	return atomic.LoadInt64(&activeTCPConnections)
}

func IncrementActiveTCPConnection() {
	atomic.AddInt64(&activeTCPConnections, 1)
}

func DecrementActiveTCPConnection() {
	atomic.AddInt64(&activeTCPConnections, -1)
}
func AddConnection(key *net.TCPConn) {
	remoteAddr := key.RemoteAddr().String()
	connection := &Connection{
		Conn:       key,
		RemoteAddr: remoteAddr,
		CreatedAt:  time.Now().Unix(),
	}
	activeConnections.Store(key, connection)
}

func RemoveConnection(key *net.TCPConn) {
	if conn, ok := activeConnections.Load(key); ok {
		conn.(*Connection).Conn.Close()
		activeConnections.Delete(key)
	}
}

func CloseAllConnections() {
	activeConnections.Range(func(key, value interface{}) bool {
		conn := key.(*Connection)
		conn.Conn.Close()
		activeConnections.Delete(key)
		return true
	})
}
