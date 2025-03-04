package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mhsantos/redis-server/internal/datastore"
	"github.com/mhsantos/redis-server/internal/protocol"
)

func init() {
	expire := expireCommand{"expire"}
	registerCommand(expire)
}

type expireCommand struct {
	name string
}

func (e expireCommand) getName() string {
	return e.name
}

func (e expireCommand) processArguments(data protocol.Array) protocol.DataType {
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
