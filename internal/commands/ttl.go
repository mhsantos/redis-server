package commands

import (
	"time"

	"github.com/mhsantos/redis-server/internal/datastore"
	"github.com/mhsantos/redis-server/internal/protocol"
)

func init() {
	ttl := ttlCommand{"ttl"}
	registerCommand(ttl)
}

type ttlCommand struct {
	name string
}

func (ttl ttlCommand) getName() string {
	return ttl.name
}

func (ttlc ttlCommand) processArguments(data protocol.Array) protocol.DataType {
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
