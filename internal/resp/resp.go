package resp

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

type DataType interface {
	Size() int
	BufferSize() int
}

type SimpleString struct {
	data string
}

type BulkString struct {
	data []byte
}

type IntegerValue struct {
	value int
}

type SimpleError struct {
	msg string
}

type Array struct {
	values []DataType
}

func (s SimpleString) Size() int {
	return len(s.data)
}

func (s SimpleString) BufferSize() int {
	return len(s.data) + 3
}

func (i IntegerValue) Size() int {
	return i.value
}

func (b BulkString) Size() int {
	return len(b.data)
}

func (b BulkString) BufferSize() int {
	sizeLength := len(strconv.Itoa(len(b.data)))
	return 1 + sizeLength + 2 + len(b.data) + 2
}

func (i IntegerValue) BufferSize() int {
	return len(strconv.Itoa(i.value)) + 3
}

func (s SimpleError) Size() int {
	return len(s.msg)
}

func (s SimpleError) BufferSize() int {
	return len(s.msg) + 3
}

func (a Array) Size() int {
	return len(a.values)
}

func (a Array) BufferSize() int {
	elementsSize := 0
	for _, element := range a.values {
		elementsSize += element.BufferSize()
	}
	sizeLength := len(strconv.Itoa(a.Size()))
	return 1 + sizeLength + 2 + elementsSize
}

func ParseFrame(buffer []byte) (DataType, error) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return nil, nil
	}
	return ParseElement(buffer)
}

func ParseElement(buffer []byte) (DataType, error) {
	switch buffer[0] {
	case '+':
		return ParseSimpleString(buffer[1:])
	case '-':
		return ParseSimpleError(buffer[1:])
	case ':':
		return ParseInteger(buffer[1:])
	case '$':
		return ParseBulkString(buffer[1:])
	case '*':
		return ParseArray(buffer[1:])
	default:
		return nil, errors.New("invalid input type")
	}
}

func ParseSimpleString(buffer []byte) (SimpleString, error) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return SimpleString{}, nil
	}
	return SimpleString{string(buffer[0:lineBreakIndex])}, nil
}

func ParseSimpleError(buffer []byte) (SimpleError, error) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return SimpleError{}, nil
	}
	return SimpleError{string(buffer[0:lineBreakIndex])}, nil
}

func ParseInteger(buffer []byte) (IntegerValue, error) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return IntegerValue{}, nil
	}
	input := string(buffer[0:lineBreakIndex])
	ival, err := strconv.Atoi(input)
	if err != nil {
		return IntegerValue{}, fmt.Errorf("error convering integer value: %s", input)
	}
	return IntegerValue{ival}, nil
}

func ParseBulkString(buffer []byte) (BulkString, error) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return BulkString{}, nil
	}
	length := string(buffer[0:lineBreakIndex])
	bulkStringLength, err := strconv.Atoi(length)
	if err != nil {
		return BulkString{}, fmt.Errorf("invalid bulk string length: %s", length)
	}
	// To account for: the initial bulk string size, the CRLF after that and the CRLF after the bulkstring
	// For example 5\r\nHello\r\n would have 5 delimiter characters: 1 + 2 + 2
	delimitersSize := lineBreakIndex + 2 + 2
	if bulkStringLength > 0 && len(buffer) >= bulkStringLength+delimitersSize {
		start := lineBreakIndex + 2
		end := start + bulkStringLength
		return BulkString{buffer[start:end]}, nil
	}
	return BulkString{}, nil
}

func ParseArray(buffer []byte) (Array, error) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return Array{}, nil
	}
	input := string(buffer[0:lineBreakIndex])
	elements, err := strconv.Atoi(input)
	if err != nil {
		return Array{}, fmt.Errorf("invalid array length: %s", input)
	}

	discardLength := lineBreakIndex + 2
	var arrayValues []DataType

	for i := 0; i < elements; i++ {
		element, err := ParseElement(buffer[discardLength:])
		if err != nil {
			return Array{}, err
		}
		if element.Size() == 0 {
			return Array{}, err
		}
		arrayValues = append(arrayValues, element)
		discardLength += element.BufferSize()
	}
	return Array{arrayValues}, nil

}
