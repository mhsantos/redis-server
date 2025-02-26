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

	// Start the task processor
	go protocol.Start()

	// Accept incoming connections and handle them
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Accepting connection from %s\n", conn.RemoteAddr())

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
	responseQueue := make(chan protocol.DataType)

	for {
		size, err := conn.Read(inBuf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Client disconnected")
				return
			}
			fmt.Println(err)
		}
		protocolBuf = append(protocolBuf, inBuf[:size]...)
		data, dataSize := protocol.ParseCommand(protocolBuf)
		if dataSize > 0 {
			// Processed a full frame
			switch data.(type) {
			case protocol.SimpleError:
				_, err = conn.Write([]byte(data.Encode()))
				clear(inBuf)
			case protocol.Array:
				task := protocol.Task{
					Command:         data.(protocol.Array),
					ResponseChannel: responseQueue,
				}
				protocol.AppendTask(task)
			}
			response := <-responseQueue
			_, err = conn.Write([]byte(response.Encode()))
			clear(inBuf)
			processedBufferSize := dataSize
			protocolBuf = protocolBuf[processedBufferSize:]
			continue
		}
		clear(inBuf)

	}

}
