package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/mhsantos/redis-server/internal/protocol"
)

const (
	bufferSize = 128
)

func main() {
	// Listen for incoming connections on port 6379
	listener, err := net.Listen("tcp", ":6379")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Accept incoming connections and handle them
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Accepting connection from %s\n", conn.RemoteAddr())

		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// Close the connection when we're done
	defer conn.Close()

	// Read incoming data
	inBuf := make([]byte, bufferSize)
	protocolBuf := make([]byte, 0)
	//outBuf := make([]byte, bufferSize)

	for {
		size, err := conn.Read(inBuf)
		fmt.Println("Bytes received", size)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Client disconnected")
				return
			}
			fmt.Println(err)
		}
		protocolBuf = append(protocolBuf, inBuf[:size]...)
		data, protocolErr := protocol.ParseFrame(protocolBuf)
		fmt.Println("data", data)
		fmt.Println("protocol", string(protocolBuf))
		if protocolErr != nil {
			_, err = conn.Write([]byte("error: " + protocolErr.Error()))
			clear(inBuf)
			continue
		}
		if data.Size() > 0 {
			// Processed a valid input
			_, err = conn.Write([]byte("Received: " + data.String()))
			clear(inBuf)
			processedBufferSize := data.BufferSize()
			protocolBuf = protocolBuf[processedBufferSize:]
			continue
		}
		clear(inBuf)

	}

}
