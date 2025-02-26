package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	store map[string]DataType = make(map[string]DataType)
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

func (i IntegerValue) String() string {
	return strconv.Itoa(i.value)
}

func (i IntegerValue) Encode() []byte {
	var buffer []byte
	buffer = append(buffer, []byte(":")...)
	buffer = append(buffer, []byte(strconv.Itoa(i.value))...)
	buffer = append(buffer, []byte("\r\n")...)
	return buffer
}

func (s SimpleError) String() string {
	return s.msg
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

// ParseCommand parses byte slice buffer input and calls the ParseFrame function to
// determine if it received a full command. If it did it will process the command returning
// a SimpleError object if the command is invalid. It always returns the number of processed
// bytes or -1 if the buffer input doesn't contain a full command.
func ParseCommand(buffer []byte) (DataType, int) {
	data, size := ParseFrame(buffer)
	if size == -1 {
		return data, -1
	}
	switch data.(type) {
	case Array:
		if len(data.(Array).values) < 1 {
			return SimpleError{fmt.Sprintf("command not informed")}, size
		}
		switch data.(Array).values[0].(type) {
		case BulkString:
			return data.(Array), size
		default:
			return SimpleError{fmt.Sprintf("invalid command of type %T. Commands must be of BulkString type", data)}, size
		}
	default:
		return SimpleError{fmt.Sprintf("invalid input of type %T. Expected an Array", data)}, size
	}
}

func processCommand(data Array) DataType {
	command := data.values[0]
	switch strings.ToLower(command.String()) {
	case "get":
		return processGet(data)
	case "set":
		return processSet(data)
	default:
		return SimpleError{fmt.Sprintf("invalid command %s", command.String())}
	}
}

func processGet(data Array) DataType {
	if len(data.values) != 2 {
		return SimpleError{fmt.Sprintf("the GET command accepts 2 parameters: GET and KEY. Received %d parameters instead", len(data.values))}
	}
	key, ok := data.values[1].(BulkString)
	if !ok {
		return SimpleError{fmt.Sprintf("the KEY parameter for the GET command must be a BulkString. Received a %T instead", data.values[1])}

	}
	val, ok := store[key.String()]
	if !ok {
		return SimpleString{"not found"}
	}
	return val
}

func processSet(data Array) DataType {
	if len(data.values) != 3 {
		return SimpleError{fmt.Sprintf("the SET command accepts 3 parameters: SET, KEY and VALUE. Received %d parameters instead", len(data.values))}
	}
	key, ok := data.values[1].(BulkString)
	if !ok {
		return SimpleError{fmt.Sprintf("the KEY parameter for the SET command must be a BulkString. Received a %T instead", data.values[1])}
	}
	store[string(key.data)] = data.values[2]
	return SimpleString{"OK"}
}
