// Package command implements the Redis commands processing logic.
//
// It makes use of the [internal/protocol/progocol.go] package to parse the input
// and get access to the actual commands that were input.

package commands

import (
	"fmt"
	"strings"

	"github.com/mhsantos/redis-server/internal/protocol"
)

var registeredCommands map[string]command

type command interface {
	getName() string
	processArguments(data protocol.Array) protocol.DataType
}

func init() {
	registeredCommands = make(map[string]command)
}

// ParseCommand parses byte slice buffer input and calls the ParseFrame function to
// determine if it received a full command. If it did it will process the command returning
// a Error object if the command is invalid. It always returns the number of processed
// bytes or -1 if the buffer input doesn't contain a full command.
func ParseCommand(buffer []byte) (protocol.DataType, int) {
	data, size := protocol.ParseFrame(buffer)
	if size == -1 {
		return data, -1
	}
	switch data.(type) {
	case protocol.Array:
		elements := data.(protocol.Array).GetElements()
		if len(elements) < 1 {
			return protocol.NewError("command not informed"), size
		}
		switch elements[0].(type) {
		case protocol.BulkString:
			return data.(protocol.Array), size
		default:
			return protocol.NewError(fmt.Sprintf("invalid command of type %T. Commands must be of BulkString type", data)), size
		}
	default:
		return protocol.NewError(fmt.Sprintf("invalid input of type %T. Expected an Array", data)), size
	}
}

//func registerCommand(process func(protocol.Array) protocol.DataType)

func registerCommand(cmd command) {
	registeredCommands[strings.ToLower(cmd.getName())] = cmd
}

func ProcessCommand(data protocol.Array) protocol.DataType {
	command := data.GetElements()[0]
	name := strings.ToLower(command.String())
	operation, ok := registeredCommands[name]
	if ok {
		return operation.processArguments(data)
	}
	fmt.Printf("invalid command %s", command.String())
	return protocol.NewError(fmt.Sprintf("invalid command %s", command.String()))
}
