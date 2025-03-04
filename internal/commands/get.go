package commands

import (
	"fmt"

	"github.com/mhsantos/redis-server/internal/datastore"
	"github.com/mhsantos/redis-server/internal/protocol"
)

func init() {
	get := getCommand{"get"}
	registerCommand(get)
}

type getCommand struct {
	name string
}

func (g getCommand) getName() string {
	return g.name
}

func (g getCommand) processArguments(data protocol.Array) protocol.DataType {
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
