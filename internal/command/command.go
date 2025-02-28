// Package command implements the Redis commands processing logic.
//
// It makes use of the [internal/protocol/progocol.go] package to parse the input
// and get access to the actual commands that were input.

package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mhsantos/redis-server/internal/datastore"
	"github.com/mhsantos/redis-server/internal/protocol"
)

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

func ProcessCommand(data protocol.Array) protocol.DataType {
	command := data.GetElements()[0]
	switch strings.ToLower(command.String()) {
	case "get":
		return processGet(data)
	case "set":
		return processSet(data)
	case "expire":
		return processExpire(data)
	case "ttl":
		return processTtl(data)
	default:
		return protocol.NewError(fmt.Sprintf("invalid command %s", command.String()))
	}
}

func processGet(data protocol.Array) protocol.DataType {
	elements := data.GetElements()
	if len(elements) != 2 {
		return protocol.NewError(fmt.Sprintf("the GET command accepts 2 parameters: GET and KEY. Received %d parameters instead", len(elements)))
	}
	key, ok := elements[1].(protocol.BulkString)
	if !ok {
		return protocol.NewError(fmt.Sprintf("the KEY parameter for the GET command must be a BulkString. Received a %T instead", elements[1]))

	}
	val, ok := datastore.Get(key.String())
	if !ok {
		return protocol.NewSimpleString("not found")
	}
	return val
}

func processSet(data protocol.Array) protocol.DataType {
	elements := data.GetElements()
	if len(elements) != 3 {
		return protocol.NewError(fmt.Sprintf("the SET command accepts 3 parameters: SET, KEY and VALUE. Received %d parameters instead", len(elements)))
	}
	key, ok := elements[1].(protocol.BulkString)
	if !ok {
		return protocol.NewError(fmt.Sprintf("the KEY parameter for the SET command must be a BulkString. Received a %T instead", elements[1]))
	}
	datastore.Set(key.String(), elements[2])
	return protocol.NewSimpleString("OK")
}

func processExpire(data protocol.Array) protocol.DataType {
	elements := data.GetElements()
	if len(elements) < 3 || len(elements) > 4 {
		return protocol.NewError(fmt.Sprintf("invalid number of arguments: %d\n", len(elements)))
	}
	key := elements[1].String()
	seconds, err := strconv.Atoi(elements[2].String())
	if err != nil {
		return protocol.NewError("seconds argument must be a positive number")
	}
	if seconds < 0 {
		datastore.Delete(key)
		return protocol.NewInteger(0)
	}
	existing, currentExpire, ok := datastore.GetWithExpire(key)
	if !ok {
		return protocol.NewInteger(0)
	}
	newExpire := time.Now().Add(time.Duration(seconds) * time.Second).Unix()
	if len(elements) == 3 {
		datastore.SetWithExpire(key, existing, newExpire)
		return protocol.NewInteger(1)
	}
	// The only remaining option is 4 arguments
	option := elements[3].String()
	switch option {
	case "NX":
		if currentExpire == 0 {
			datastore.SetWithExpire(key, existing, newExpire)
			return protocol.NewInteger(1)
		}
	case "XX":
		if currentExpire > 0 {
			datastore.SetWithExpire(key, existing, newExpire)
			return protocol.NewInteger(1)
		}
	case "GT":
		if currentExpire > 0 && newExpire > currentExpire {
			datastore.SetWithExpire(key, existing, newExpire)
			return protocol.NewInteger(1)
		}
	case "LT":
		if currentExpire > 0 && newExpire < currentExpire {
			datastore.SetWithExpire(key, existing, newExpire)
			return protocol.NewInteger(1)
		}
	default:
		return protocol.NewError(fmt.Sprintf("invalid option %s\n", option))
	}
	return protocol.NewInteger(0)
}

func processTtl(data protocol.Array) protocol.DataType {
	elements := data.GetElements()
	if len(elements) != 2 {
		return protocol.NewError("invalid arguments")
	}
	key := elements[1].String()
	_, currentExpire, ok := datastore.GetWithExpire(key)
	if !ok {
		return protocol.NewInteger(-2)
	}
	if currentExpire == 0 {
		return protocol.NewInteger(-1)
	}
	ttl := currentExpire - time.Now().Unix()
	return protocol.NewInteger(int(ttl))
}
