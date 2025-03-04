package commands

import (
	"fmt"

	"github.com/mhsantos/redis-server/internal/datastore"
	"github.com/mhsantos/redis-server/internal/protocol"
)

func init() {
	exists := existsCommand{"exists"}
	registerCommand(exists)
}

type existsCommand struct {
	name string
}

func (e existsCommand) getName() string {
	return e.name
}

func (e existsCommand) processArguments(data protocol.Array) protocol.DataType {
	elements := data.GetElements()
	if len(elements) < 2 {
		return protocol.NewError("invalid arguments for command EXPIRE. Syntax: EXPIRE key1 [key2] [key3] ...")
	}
	sum := 0
	for _, element := range elements[1:] {
		key, ok := element.(protocol.BulkString)
		if !ok {
			return protocol.NewError(fmt.Sprintf("the KEY parameter for the EXISTS command must be a BulkString. Received a %T instead", element))

		}
		if _, ok := datastore.Get(key.String()); ok {
			sum++
		}
	}
	return protocol.NewInteger(sum)
}
