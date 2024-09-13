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
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"universum/config"
	"universum/engine"
	"universum/entity"
	"universum/internal/logger"
	"universum/resp3"
)

const networkTcp string = "tcp"

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
		atomic.StoreInt32(&entity.ServerState, entity.STATE_SHUTTING_DOWN)
	}()

	atomic.StoreInt32(&entity.ServerState, entity.STATE_STARTING)

	port := fmt.Sprintf(":%d", config.GetServerPort())
	maxConnections := config.GetMaxClientConnections()

	// Create a channel to manage a pool of connections
	connectionQueue := make(chan *net.TCPConn, maxConnections)

	// Start a worker pool for processing connections
	for i := 0; i < int(maxConnections); i++ {
		go connectionWorker(i+1, connectionQueue)
	}

	// Listen on the configured port
	listener, err := net.Listen(networkTcp, port)
	if err != nil {
		logger.Get().Error("Error listening on socket, will shutdown: %v", err.Error())
		engine.Shutdown(entity.ExitCodeSocketError)
	}

	defer listener.Close()
	logger.Get().Info("%s server started listening on port %s", config.AppCodeName, port)

	engine.Startup()
	atomic.StoreInt32(&entity.ServerState, entity.STATE_READY)

	for {
		// If the server is shutting down, handle remaining requests and stop accepting new ones
		if atomic.LoadInt32(&entity.ServerState) == entity.STATE_SHUTTING_DOWN {
			go handleRequestWhenShuttingDown(listener)
			continue
		}

		// Accept new incoming connections
		conn, err := listener.Accept()
		if err != nil {
			logger.Get().Error("Error accepting client connection: %v", err.Error())
			continue
		}

		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			logger.Get().Error("Invalid connection type: Non-TCP connection detected")
			conn.Close()
			continue
		}

		// Add connection tracking and increase the active connection count
		entity.AddConnection(tcpConn)
		entity.IncrementActiveTCPConnection()

		// Set NoDelay to prevent Nagle's algorithm, ensuring faster response
		tcpConn.SetNoDelay(true)

		// Limit total accepted connections, dropping ones when max connections reached
		select {
		case connectionQueue <- tcpConn:
			// Successfully added connection to the worker queue
			continue
		default:
		case <-time.After(100 * time.Millisecond): // Retry timeout
			logger.Get().Warn("Server busy, dropping incoming connection")
			closeTCPConnection(conn) // Close the connection gracefully after timeout
		}
	}
}

// connectionWorker represents a worker in the server's worker pool. Each worker
// listens for incoming jobs on the jobs channel, processes connections, and
// handles concurrency based on the number of connections.
//
// Parameters:
// - id int: The unique ID of the worker (for logging and tracking purposes).
// - queue <-chan *net.TCPConn: A channel of TCP connections to be processed by the worker.
func connectionWorker(id int, queue <-chan *net.TCPConn) {
	defer func() {
		// Panic recovery to keep the worker alive in case of runtime errors
		if r := recover(); r != nil {
			logger.Get().Error("Connection worker [id=%d] panicked: %v. Restarting...", id, r)
			go connectionWorker(id, queue)
		}
	}()

	for conn := range queue {
		handleConnection(conn)
	}
}

// handleConnection is responsible for the lifecycle of a single TCP connection
// from acceptance through processing of commands and closure. It reads commands
// from the connection, executes them using the engine, and writes responses back
// to the client. It ensures that each connection is processed efficiently and
// safely.
//
// Parameters:
// - conn *net.TCPConn: The TCP connection to be handled.
func handleConnection(conn *net.TCPConn) {
	buffer := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Recover from panics and close the connection gracefully
	defer func() {
		err := errors.New("connection pipe broken, closing the connection")

		outputWithEOM := resp3.EncodedRESP3Response(err) + entity.ResponseDelimiter
		engine.AddNetworkBytesSent(int64(len(outputWithEOM)))

		writer.Write([]byte(outputWithEOM))
		writer.Flush()

		closeTCPConnection(conn)
	}()

	reqTimeout := config.GetRequestExecutionTimeout()
	writeTimeout := config.GetTCPConnectionWriteTimeout()

	for {
		// Execute the client command with a request timeout
		output, err := engine.ExecuteCommand(buffer, reqTimeout)

		if err != nil {
			if err == io.EOF {
				// Connection closed by the client
				return
			}
			// Handle other errors
			output = resp3.EncodedRESP3Response(err)
		}

		outputWithEOM := output + entity.ResponseDelimiter

		// Set a write deadline for sending the response
		conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		_, err = writer.Write([]byte(outputWithEOM))

		if err != nil {
			// Handle socket write errors and terminate the connection if needed
			if _, ok := err.(net.Error); ok {
				logger.Get().Debug("Connection timed out or dropped, closing the connection")
				return
			}
			logger.Get().Error("Error writing to the socket: %v", err.Error())
		}

		engine.AddNetworkBytesSent(int64(len(outputWithEOM)))

		flushErr := writer.Flush()
		if flushErr != nil {
			logger.Get().Debug("Unable to flush the response: %v", flushErr)
			return
		}
	}
}

// handleRequestWhenShuttingDown handles new connections while the server is
// shutting down. It accepts a new connection and sends a shutdown message to the
// client, indicating that the server cannot process new requests.
//
// Parameters:
// - listener net.Listener: The server's listener to accept new connections.
func handleRequestWhenShuttingDown(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		logger.Get().Error("Error accepting connection during shutdown: %v", err.Error())
		return
	}

	entity.IncrementActiveTCPConnection()

	shutdownMessage := getFormattedClientMessage(nil,
		entity.CRC_SERVER_SHUTTING_DOWN,
		"Server shutting down, cannot serve the request.",
	)

	outputWithEOM := resp3.EncodedRESP3Response(shutdownMessage) + entity.ResponseDelimiter

	writer := bufio.NewWriter(conn)
	writer.Write([]byte(outputWithEOM))
	writer.Flush()

	closeTCPConnection(conn)
}

// WaitForSignal blocks until a shutdown signal is received. It initiates a
// graceful shutdown process, ensuring that the server stops accepting new
// connections and properly terminates existing ones.
//
// Parameters:
// - wg *sync.WaitGroup: A WaitGroup to manage the shutdown goroutine's lifecycle.
// - sigs chan os.Signal: A channel to receive operating system signals.
func WaitForSignal(wg *sync.WaitGroup, sigs chan os.Signal) {
	defer wg.Done()
	receivedSignal := <-sigs

	logger.Get().Fatal("Shutting down the server due to signal: %s", receivedSignal.String())

	// Set the state to SHUTTING DOWN to stop accepting new connections
	atomic.StoreInt32(&entity.ServerState, entity.STATE_SHUTTING_DOWN)

	engine.Shutdown(entity.ExitCodeInturrupted)
}

// closeTCPConnection closes a TCP connection and removes it from the active
// connections list while decrementing the active connection count.
//
// Parameters:
// - conn net.Conn: The TCP connection to be closed.
func closeTCPConnection(conn net.Conn) {
	conn.Close()
	entity.RemoveConnection(conn.(*net.TCPConn))
	entity.DecrementActiveTCPConnection()
}
