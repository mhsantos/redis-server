package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/mhsantos/redis-server/internal/command"
	"github.com/mhsantos/redis-server/internal/protocol"
	"github.com/mhsantos/redis-server/internal/taskmanager"
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
	go taskmanager.Start()

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
	defer func() {
		r := recover()
		fmt.Printf("error parsing input %s. closing the connection\n", r.(error))
		// Close the connection when we're done
		conn.Close()
	}()

	// Read incoming data
	inBuf := make([]byte, bufferSize)
	protocolBuf := make([]byte, 0)
	responseQueue := make(chan protocol.DataType)

	for {
		size, err := conn.Read(inBuf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected")
				return
			} else {
				fmt.Println(err)
			}
		}
		protocolBuf = append(protocolBuf, inBuf[:size]...)
		data, dataSize := command.ParseCommand(protocolBuf)
		if dataSize > 0 {
			// Processed a full frame
			switch data := data.(type) {
			case protocol.Error:
				conn.Write([]byte(data.Encode()))
				clear(inBuf)
			case protocol.Array:
				task := taskmanager.Task{
					Command:         data,
					ResponseChannel: responseQueue,
				}
				taskmanager.AppendTask(task)
			}
			response := <-responseQueue
			conn.Write([]byte(response.Encode()))
			clear(inBuf)
			processedBufferSize := dataSize
			protocolBuf = protocolBuf[processedBufferSize:]
			continue
		}
		clear(inBuf)
	}

}
