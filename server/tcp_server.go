// Package server implements a TCP server for handling concurrent client
// connections in a scalable and efficient manner. It's designed to support
// an in-memory key-value pair database, offering functionalities such as
// connection throttling, state management, and graceful shutdown.
//
// The server manages its lifecycle states, including starting, ready, busy,
// and shutting down, to efficiently handle incoming connections based on
// the current load and system resources. It employs a worker pool pattern
// to distribute connection processing across multiple goroutines, ensuring
// responsive and concurrent handling of client requests.
package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"universum/config"
	"universum/consts"
	"universum/engine"
	"universum/internal/logger"
	"universum/resp3"
)

var maxRetryCountWhenBusy int = 5
var serverState int32

// StartTCPServer initializes and starts the TCP server, managing incoming
// client connections. It sets up a listener on the configured port, utilizes
// a worker pool to process connections concurrently, and manages server
// states to optimize for load and performance. The function should be run
// in a goroutine and is intended to block until server shutdown is initiated.
//
// Parameters:
// - wg *sync.WaitGroup: A WaitGroup to manage the lifecycle of server goroutines.
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
		logger.Get().Error("Error listening on socket: %v", err.Error())
		os.Exit(1)
	}

	defer listener.Close()
	logger.Get().Info("%s server started listening on port %s", config.APP_CODE_NAME, port)

	engine.Startup()
	atomic.StoreInt32(&serverState, STATE_READY)

	for {
		if atomic.LoadInt32(&serverState) == STATE_SHUTTING_DOWN {
			handleRequestWhenShuttingDown(listener)
		}

		if atomic.LoadInt32(&serverState) == STATE_BUSY {
			// server is potentially overloaded with requests, so wait for some time.
			for i := 0; i < maxRetryCountWhenBusy; i++ {
				if atomic.LoadInt32(&serverState) == STATE_READY {
					break
				}

				time.Sleep(100 * time.Millisecond)
			}

			if atomic.LoadInt32(&serverState) != STATE_READY {
				// if server is still not ready, means its not wise
				// to wait for longer. so do something to close the connection
				handleRequestWhenServerBusy(listener)
			}
		}

		conn, err := listener.Accept()
		tcpConn, _ := conn.(*net.TCPConn)

		if err != nil {
			logger.Get().Error("Error accepting client connection: %v", err.Error())
			continue
		}

		// set nodelay=true for tcp connection, so the response is immediately
		// sent to the client without waiting lazily for response because of nagle's algorithm
		tcpConn.SetNoDelay(true)

		// Limit total accepted connections by trying to insert into the limiter channel.
		// If it's full, we've reached the max and should handle this scenario.
		select {
		case connectionLimiter <- struct{}{}:
			// We successfully inserted a token, meaning we haven't reached the max.
			jobs <- tcpConn

		default:
			// Reached max connections; change the server state to busy, if not already
			// set and refuse the current incoming request.
			if atomic.LoadInt32(&serverState) == STATE_READY {
				atomic.StoreInt32(&serverState, STATE_BUSY)
			}

			logger.Get().Warn("Max connections reached, refusing this connection")
			conn.Close()
		}
	}

}

// concurrentWorker represents a worker in the server's worker pool. Each worker
// listens for incoming jobs on the jobs channel, processes connections, and
// adheres to the concurrency limit defined by the connection limiter. This
// function is designed to run as a goroutine, processing connections in a loop
// until the server is shut down.
//
// Parameters:
// - jobs <-chan *net.TCPConn: A channel of TCP connections to be processed by the worker.
// - connectionLimiter <-chan struct{}: A channel used to control the rate of connection processing.
func concurrentWorker(jobs <-chan *net.TCPConn, connectionLimiter <-chan struct{}) {
	for conn := range jobs {
		handleConnection(conn)
		// After handling, release a spot in the limiter.
		<-connectionLimiter
		atomic.StoreInt32(&serverState, STATE_READY)
	}
}

// handleConnection is responsible for the lifecycle of a single TCP connection
// from acceptance through processing of commands and closure. It reads commands
// from the connection, executes them using the engine, and writes responses back
// to the client. It ensures that each connection is processed efficiently and
// safely, implementing read deadlines and recovery from panics to maintain server
// stability.
//
// Parameters:
// - conn *net.TCPConn: The TCP connection to be handled.
func handleConnection(conn *net.TCPConn) {
	buffer := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// recover from the panic if any faulty client may cause
	// any kind of trouble to the server
	defer func() {
		err := errors.New("connection pipe broken, closing the connection")

		atomic.StoreInt32(&serverState, STATE_READY)

		outputWithEOM := resp3.EncodedRESP3Response(err) + consts.RESPONSE_DELIMITER
		writer.Write([]byte(outputWithEOM))
		writer.Flush()
		conn.Close()
	}()

	readtimeout := config.GetTCPConnectionReadtime()

	for {
		conn.SetReadDeadline(time.Now().Add(readtimeout))
		output, err := engine.ExecuteCommand(buffer)

		if err != nil {
			output = resp3.EncodedRESP3Response(err)
		}

		outputWithEOM := output + consts.RESPONSE_DELIMITER
		_, err = writer.Write([]byte(outputWithEOM))

		if err != nil {
			// of connection has timed out or dropped, then terminate the flow
			if _, ok := err.(net.Error); ok {
				return
			}

			logger.Get().Error("Error writing to the socket: %v", err.Error())
		}

		writer.Flush()
	}
}

// handleRequestWhenShuttingDown accepts a new connection and sends a shutdown
// message to the client, indicating that the server is currently shutting down
// and cannot process the request. This function is intended to be called when
// the server is in the process of shutting down but still accepting connections
// to provide feedback to clients.
//
// Parameters:
// - listener net.Listener: The server's listener to accept new connections.
func handleRequestWhenShuttingDown(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		logger.Get().Error("Error accepting connection during shutdown: %v", err.Error())
		return
	}

	shutdownMessage := getFormattedClientMessage(nil,
		consts.CRC_SERVER_SHUTTING_DOWN,
		"Server shutting down, cannot serve the request.",
	)

	outputWithEOM := resp3.EncodedRESP3Response(shutdownMessage) + consts.RESPONSE_DELIMITER

	writer := bufio.NewWriter(conn)
	writer.Write([]byte(outputWithEOM))
	writer.Flush()
	conn.Close()
}

// handleRequestWhenServerBusy accepts a new connection and sends a busy message
// to the client, advising them that the server is currently overloaded and
// suggesting a retry after some backoff. This function is used to manage
// incoming connections when the server has reached its concurrency limit,
// improving client experience by providing actionable feedback.
//
// Parameters:
// - listener net.Listener: The server's listener to accept new connections.
func handleRequestWhenServerBusy(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		logger.Get().Error("Error accepting connection while server busy: %v", err.Error())
		return
	}

	shutdownMessage := getFormattedClientMessage(nil,
		consts.CRC_SERVER_BUSY,
		"Server is busy, try connecting after some backoff.",
	)

	outputWithEOM := resp3.EncodedRESP3Response(shutdownMessage) + consts.RESPONSE_DELIMITER

	writer := bufio.NewWriter(conn)
	writer.Write([]byte(outputWithEOM))
	writer.Flush()
	conn.Close()
}

// WaitForSignal blocks until a shutdown signal is received. It then initiates
// a graceful shutdown process, ensuring that the server ceases to accept new
// connections and properly terminates existing connections before exiting.
// This function should be run in a goroutine and is responsible for managing
// the server's shutdown in response to external signals.
//
// Parameters:
// - wg *sync.WaitGroup: A WaitGroup to manage the lifecycle of the shutdown goroutine.
// - sigs chan os.Signal: A channel to receive operating system signals.
func WaitForSignal(wg *sync.WaitGroup, sigs chan os.Signal) {
	defer wg.Done()
	receivedSignal := <-sigs

	logger.Get().Fatal("Shutting down the server due to signal: %s", receivedSignal.String())

	// if server is busy continue to wait with period sleep of 100ms
	for atomic.LoadInt32(&serverState) == STATE_BUSY {
		time.Sleep(time.Millisecond * 100)
	}

	// immediately set the status to be SHUTTING DOWN,
	// so it does not start taking more connections.
	atomic.StoreInt32(&serverState, STATE_SHUTTING_DOWN)

	engine.Shutdown()
}
