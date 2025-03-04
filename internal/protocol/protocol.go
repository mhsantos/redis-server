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

type Integer struct {
	value int
}

type Error struct {
	msg string
}

type Array struct {
	elements []DataType
}

func (s SimpleString) String() string {
	return s.data
}

func (s SimpleString) Encode() []byte {
	var buffer []byte
	buffer = append(buffer, []byte("+")...)
	buffer = append(buffer, []byte(s.data)...)
	buffer = append(buffer, []byte("\r\n")...)
	return buffer
}

func (b BulkString) String() string {
	return string(b.data)
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

func (i Integer) String() string {
	return strconv.Itoa(i.value)
}

func (i Integer) Encode() []byte {
	var buffer []byte
	buffer = append(buffer, []byte(":")...)
	buffer = append(buffer, []byte(strconv.Itoa(i.value))...)
	buffer = append(buffer, []byte("\r\n")...)
	return buffer
}

func (s Error) String() string {
	return s.msg
}

func (s Error) Encode() []byte {
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
	for _, val := range a.elements {
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
	buffer = append(buffer, []byte(strconv.Itoa(len(a.elements)))...)
	buffer = append(buffer, []byte("\r\n")...)
	for _, val := range a.elements {
		buffer = append(buffer, val.Encode()...)
	}
	return buffer
}

func (a Array) GetElements() []DataType {
	return a.elements
}

func NewBulkString(value []byte) BulkString {
	return BulkString{value}
}

func NewSimpleString(value string) SimpleString {
	return SimpleString{value}
}

func NewError(msg string) Error {
	return Error{msg}
}

func NewInteger(value int) Integer {
	return Integer{value}
}

func NewArray(elements ...DataType) Array {
	return Array{elements}
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
	return parseElement(buffer)
}

func parseElement(buffer []byte) (DataType, int) {
	switch buffer[0] {
	case '+':
		return ParseSimpleString(buffer[1:])
	case '-':
		return ParseError(buffer[1:])
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

func ParseError(buffer []byte) (Error, int) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return Error{}, -1
	}
	return Error{string(buffer[0:lineBreakIndex])}, 1 + lineBreakIndex + 2
}

func ParseInteger(buffer []byte) (Integer, int) {
	lineBreakIndex := bytes.Index(buffer, []byte("\r\n"))
	if lineBreakIndex == -1 {
		return Integer{}, -1
	}
	input := string(buffer[0:lineBreakIndex])
	ival, err := strconv.Atoi(input)
	if err != nil {
		panic(fmt.Errorf("error reading integer %s: %w", input, err))
	}
	return Integer{ival}, 1 + lineBreakIndex + 2
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
		element, byteSize := parseElement(buffer[bytesRead:])
		if byteSize < 0 {
			return Array{}, -1
		}
		bytesRead += byteSize
		arrayValues = append(arrayValues, element)
	}
	return Array{arrayValues}, 1 + bytesRead
}
