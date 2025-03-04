package commands

import (
	"fmt"

	"github.com/mhsantos/redis-server/internal/datastore"
	"github.com/mhsantos/redis-server/internal/protocol"
)

const (
	delInvalidLengthErrMsg string = "invalid arguments for command DEL. Syntax: DEL key1 [key2] [key3] ..."
	delKeyTypeErrMsg       string = "the KEY parameter for the DEL command must be a BulkString. Received a %T instead"
)

func init() {
	del := delCommand{
		name: "del",
	}
	registerCommand(del)
}

type delCommand struct {
	name string
}

func (d delCommand) getName() string {
	return d.name
}

func (d delCommand) processArguments(data protocol.Array) protocol.DataType {
	elements := data.GetElements()
	if len(elements) < 2 {
		return protocol.NewError(delInvalidLengthErrMsg)
	}
	deleted := 0
	for _, element := range elements[1:] {
		key, ok := element.(protocol.BulkString)
		if !ok {
			return protocol.NewError(fmt.Sprintf(delKeyTypeErrMsg, element))

		}
		if ok := datastore.Delete(key.String()); ok {
			deleted++
		}
	}
	return protocol.NewInteger(deleted)
}
