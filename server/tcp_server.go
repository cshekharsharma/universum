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

	jobs := make(chan net.Conn, concurrencyLimit)
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
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		// Limit total accepted connections by trying to insert into the limiter channel.
		// If it's full, we've reached the max and should handle this scenario.
		select {
		case connectionLimiter <- struct{}{}:
			// We successfully inserted a token, meaning we haven't reached the max.
			// Set a read deadline and enqueue the connection.
			conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
			jobs <- conn
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

func concurrentWorker(jobs <-chan net.Conn, connectionLimiter <-chan struct{}) {
	for conn := range jobs {
		handleConnection(conn)
		// After handling, release a spot in the limiter.
		<-connectionLimiter
		atomic.StoreInt32(&serverState, STATE_READY)
	}
}

func handleConnection(conn net.Conn) {
	buffer := bufio.NewReader(conn)

	output, err := engine.ExecuteCommand(buffer)
	fmt.Printf("THREYOUGO: %#v\n", output)

	if err != nil {
		output = engine.EncodedResponse(err)
	}

	_, err = conn.Write([]byte(output))
	if err != nil {
		fmt.Printf("Error writing to the socket: %v\n", err)
	}

	conn.Close()
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
