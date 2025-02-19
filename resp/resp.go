package resp

import (
	"bytes"
	"errors"
)

type SimpleString struct {
	data string
}

func ParseFrame(buffer []byte) (SimpleString, error) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return SimpleString{}, nil
	}
	switch buffer[0] {
	case '+':
		return SimpleString{
			data: string(buffer[1:lineBreakIndex]),
		}, nil

	case '-':
		return SimpleString{}, errors.New(string(buffer[1:lineBreakIndex]))
	default:
		return SimpleString{}, nil
	}
}
