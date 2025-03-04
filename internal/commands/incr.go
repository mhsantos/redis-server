package commands

import (
	"fmt"
	"strconv"

	"github.com/mhsantos/redis-server/internal/datastore"
	"github.com/mhsantos/redis-server/internal/protocol"
)

const (
	incrInvalidLengthErrMsg string = "the INCR command accepts 2 parameters: INCR and KEY. Received %d parameters instead"
	incrKeyTypeErrMsg       string = "the KEY parameter for the INCR command must be a BulkString. Received a %T instead"
)

func init() {
	incr := incrCommand{"incr"}
	registerCommand(incr)
}

type incrCommand struct {
	name string
}

func (g incrCommand) getName() string {
	return g.name
}

func (g incrCommand) processArguments(data protocol.Array) protocol.DataType {
	elements := data.GetElements()
	if len(elements) != 2 {
		return protocol.NewError(fmt.Sprintf(incrInvalidLengthErrMsg, len(elements)))
	}
	key, ok := elements[1].(protocol.BulkString)
	if !ok {
		return protocol.NewError(fmt.Sprintf(incrKeyTypeErrMsg, elements[1]))

	}
	val, ok := datastore.Get(key.String())
	if !ok {
		val = protocol.NewSimpleString("0")
		datastore.Set(key.String(), val)
	}
	asNumber, err := strconv.Atoi(val.String())
	if err != nil {
		return protocol.NewError("value can't be converted to number")
	}
	asNumber++
	asStr := strconv.Itoa(asNumber)
	datastore.Set(key.String(), protocol.NewSimpleString(asStr))
	return protocol.NewInteger(asNumber)
}
