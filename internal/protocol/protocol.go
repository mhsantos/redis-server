package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type DataType interface {
	String() string
	Encode() []byte
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

func (s SimpleString) String() string {
	return fmt.Sprintf("SimpleString(%s)", s.data)
}

func (s SimpleString) Encode() []byte {
	var buffer []byte
	buffer = append(buffer, []byte("+")...)
	buffer = append(buffer, []byte(s.data)...)
	buffer = append(buffer, []byte("\r\n")...)
	return buffer
}

func (b BulkString) String() string {
	return fmt.Sprintf("BulkString(%s)", b.data)
}

func (b BulkString) Encode() []byte {
	var buffer []byte
	buffer = append(buffer, []byte("$")...)
	buffer = append(buffer, []byte(strconv.Itoa(len(b.data)))...)
	buffer = append(buffer, []byte("\r\n")...)
	buffer = append(buffer, []byte(b.data)...)
	buffer = append(buffer, []byte("\r\n")...)
	return buffer
}

func (i IntegerValue) String() string {
	return fmt.Sprintf("IntegerValue(%d)", i.value)
}

func (i IntegerValue) Encode() []byte {
	var buffer []byte
	buffer = append(buffer, []byte(":")...)
	buffer = append(buffer, []byte(strconv.Itoa(i.value))...)
	buffer = append(buffer, []byte("\r\n")...)
	return buffer
}

func (s SimpleError) String() string {
	return fmt.Sprintf("SimpleError(%s)", s.msg)
}

func (s SimpleError) Encode() []byte {
	var buffer []byte
	buffer = append(buffer, []byte("-")...)
	buffer = append(buffer, []byte(s.msg)...)
	buffer = append(buffer, []byte("\r\n")...)
	return buffer
}

func (a Array) String() string {
	builder := new(strings.Builder)
	builder.WriteString("Array[")
	elements := []string{}
	for _, val := range a.values {
		elements = append(elements, val.String())
	}
	joinedElements := strings.Join(elements, ",")
	builder.WriteString(joinedElements)
	builder.WriteString("]")
	return builder.String()
}

func (a Array) Encode() []byte {
	buffer := []byte{}
	buffer = append(buffer, []byte("*")...)
	buffer = append(buffer, []byte(strconv.Itoa(len(a.values)))...)
	buffer = append(buffer, []byte("\r\n")...)
	for _, val := range a.values {
		buffer = append(buffer, val.Encode()...)
	}
	return buffer
}

/* ParseFrame parses the buffer input. It it has a complete message, it returs the appropriate
 * DataType implementation with the number of bytes read. If it doesn't have a complete
 * input message, returns nil and -1
 */
func ParseFrame(buffer []byte) (DataType, int) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return nil, -1
	}
	return ParseElement(buffer)
}

func ParseElement(buffer []byte) (DataType, int) {
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
		panic(errors.New("invalid input type"))
	}
}

func ParseSimpleString(buffer []byte) (SimpleString, int) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return SimpleString{}, -1
	}
	return SimpleString{string(buffer[0:lineBreakIndex])}, 1 + lineBreakIndex + 2
}

func ParseSimpleError(buffer []byte) (SimpleError, int) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return SimpleError{}, -1
	}
	return SimpleError{string(buffer[0:lineBreakIndex])}, 1 + lineBreakIndex + 2
}

func ParseInteger(buffer []byte) (IntegerValue, int) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return IntegerValue{}, -1
	}
	input := string(buffer[0:lineBreakIndex])
	ival, err := strconv.Atoi(input)
	if err != nil {
		panic(fmt.Errorf("error reading integer %s: %w", input, err))
	}
	return IntegerValue{ival}, 1 + lineBreakIndex + 2
}

func ParseBulkString(buffer []byte) (BulkString, int) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return BulkString{}, -1
	}
	length := string(buffer[0:lineBreakIndex])
	bulkStringLength, err := strconv.Atoi(length)
	if err != nil {
		panic(fmt.Errorf("invalid bulk string length %s: %w", length, err))
	}
	// To account for: the initial bulk string size, the CRLF after that and the CRLF after the bulkstring
	// For example 5\r\nHello\r\n would have 5 delimiter characters: 1 + 2 + 2
	delimitersSize := lineBreakIndex + 2 + 2
	if bulkStringLength > 0 && len(buffer) >= bulkStringLength+delimitersSize {
		start := lineBreakIndex + 2
		end := start + bulkStringLength
		return BulkString{buffer[start:end]}, 1 + bulkStringLength + delimitersSize
	}
	return BulkString{}, -1
}

func ParseArray(buffer []byte) (Array, int) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return Array{}, -1
	}
	input := string(buffer[0:lineBreakIndex])
	elements, err := strconv.Atoi(input)
	if err != nil {
		panic(fmt.Errorf("invalid array length: %s", input))
	}

	bytesRead := lineBreakIndex + 2
	var arrayValues []DataType

	for i := 0; i < elements; i++ {
		if len(buffer) <= bytesRead {
			return Array{}, -1
		}
		element, byteSize := ParseElement(buffer[bytesRead:])
		if byteSize < 0 {
			return Array{}, -1
		}
		bytesRead += byteSize
		arrayValues = append(arrayValues, element)
	}
	return Array{arrayValues}, 1 + bytesRead

}
