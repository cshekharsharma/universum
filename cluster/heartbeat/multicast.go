package heartbeat

import (
	"fmt"
	"net"
	"time"
)

func StartMulticastHeartbeat(address string, port int64) error {
	multicastAddr := fmt.Sprintf("%s:%d", address, port)
	addr, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		message := "Node is alive"
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending multicast heartbeat:", err)
		}
		fmt.Println("Sent multicast heartbeat to", multicastAddr)
	}
	return nil
}
