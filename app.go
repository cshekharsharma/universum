package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

func main() {
    // Listen on TCP port 3001.
    listener, err := net.Listen("tcp", ":3001")
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    // Close the listener when the application closes.
    defer listener.Close()
    fmt.Println("Listening on 0.0.0.0:3001")

    for {
        // Accept a connection.
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }

        // Handle the connection in a new goroutine.
        // The loop then returns to accepting, so that
        // multiple connections may be served concurrently.
        go handleRequest(conn)
    }
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
    // Make a buffer to hold incoming data.
    buf := make([]byte, 1024)
    // Read the incoming connection into the buffer.
    len, err := conn.Read(buf)
    if err != nil {
        fmt.Println("Error reading:", err.Error())
    }
    // Send a response back to person contacting us.
    conn.Write(buf[:len])
    // Close the connection when you're done with it.
    conn.Close()
}
