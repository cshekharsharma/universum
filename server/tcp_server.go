package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"universum/config"
	"universum/engine"
)

var serverState int32

func StartTCPServer(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		atomic.StoreInt32(&serverState, STATE_SHUTTING_DOWN)
	}()

	atomic.StoreInt32(&serverState, STATE_STARTING)

	port := fmt.Sprintf(":%d", config.GetServerPort())
	maxConnections := config.GetMaxClientConnections()
	concurrencyLimit := config.GetServerConcurrencyLimit(maxConnections)

	jobs := make(chan *net.TCPConn, concurrencyLimit)
	connectionLimiter := make(chan struct{}, maxConnections-concurrencyLimit)

	for i := 0; i < concurrencyLimit; i++ {
		go concurrentWorker(jobs, connectionLimiter)
	}

	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error listening on socket:", err.Error())
		os.Exit(1)
	}

	defer listener.Close()
	log.Printf("server listening on port %s\n", port)

	engine.Startup()
	atomic.StoreInt32(&serverState, STATE_READY)

	for {
		conn, err := listener.Accept()
		tcpConn, _ := conn.(*net.TCPConn)

		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		// set nodelay=true, so the response is immediately sent to buffer
		// without waiting because of nagle's algorithm
		tcpConn.SetNoDelay(true)

		// Limit total accepted connections by trying to insert into the limiter channel.
		// If it's full, we've reached the max and should handle this scenario.
		select {
		case connectionLimiter <- struct{}{}:
			// We successfully inserted a token, meaning we haven't reached the max.
			// Set a read deadline and enqueue the connection.
			conn.SetReadDeadline(time.Now().Add(config.GetTCPConnectionReadtime()))
			jobs <- tcpConn
		default:
			// Reached max connections; change the server state to busy, if not already
			// set and refuse the current incoming request.
			if atomic.LoadInt32(&serverState) == STATE_READY {
				atomic.StoreInt32(&serverState, STATE_BUSY)
			}

			fmt.Println("Max connections reached, refusing connection.")
			conn.Close()
		}
	}

}

func concurrentWorker(jobs <-chan *net.TCPConn, connectionLimiter <-chan struct{}) {
	for conn := range jobs {
		handleConnection(conn)
		// After handling, release a spot in the limiter.
		<-connectionLimiter
		atomic.StoreInt32(&serverState, STATE_READY)
	}
}

func handleConnection(conn *net.TCPConn) {
	buffer := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	writer.Buffered()
	defer conn.Close()

	readtimeout := config.GetTCPConnectionReadtime()

	for {
		conn.SetReadDeadline(time.Now().Add(readtimeout))
		output, err := engine.ExecuteCommand(buffer)

		if output != "" {
			fmt.Printf("RESPONSE: %#v\n", output)
		}

		if err != nil {
			// if connection has timed out or dropped, then terminate the flow
			if _, ok := err.(net.Error); ok {
				return
			}

			output = engine.EncodedResponse(err)
		}

		outputWithEOM := []byte(output)

		for i1 := 0; i1 < 4; i1++ {
			outputWithEOM = append(outputWithEOM, 0x04)
		}
		_, err = writer.Write([]byte(outputWithEOM))

		if err != nil {
			// of connection has timed out or dropped, then terminate the flow
			if _, ok := err.(net.Error); ok {
				return
			}

			fmt.Printf("Error writing to the socket: %v\n", err)
		}

		writer.Flush()
	}
}

func WaitForSignal(wg *sync.WaitGroup, sigs chan os.Signal) {
	defer wg.Done()
	<-sigs

	// if server is busy continue to wait with period sleep of 100ms
	for atomic.LoadInt32(&serverState) == STATE_BUSY {
		time.Sleep(time.Millisecond * 100)
	}

	// immediately set the status to be SHUTTING DOWN,
	// so it does not start taking more connections.
	atomic.StoreInt32(&serverState, STATE_SHUTTING_DOWN)

	engine.Shutdown()
	os.Exit(0)
}
