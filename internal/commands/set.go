package commands

import (
	"fmt"

	"github.com/mhsantos/redis-server/internal/datastore"
	"github.com/mhsantos/redis-server/internal/protocol"
)

func init() {
	set := setCommand{"set"}
	registerCommand(set)
}

type setCommand struct {
	name string
}

func (s setCommand) getName() string {
	return s.name
}

func (s setCommand) processArguments(data protocol.Array) protocol.DataType {
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
