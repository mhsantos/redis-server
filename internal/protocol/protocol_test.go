/*
# Redis Protocol:
# Spec:

# For Simple Strings, the first byte of the reply is "+"     "+OK\r\n"
# For Errors, the first byte of the reply is "-"             "-Error message\r\n"
# For Integers, the first byte of the reply is ":"           ":[<+|->]<value>\r\n"
# For Bulk Strings, the first byte of the reply is "$"       "$<length>\r\n<data>\r\n"
# For Arrays, the first byte of the reply is "*"             "*<number-of-elements>\r\n<element-1>...<element-n>"

# We will need a module to extract messages from the stream.
# When we read from the network we will get:
# 1. A partial message.
# 2. A whole message.
# 3. A whole message, followed by either 1 or 2.
# We will need to remove parsed bytes from the stream.
*/

package protocol

import (
	"reflect"
	"testing"
)

type testCase struct {
	name     string
	input    string
	expected any
	length   int
}

func TestPartial(t *testing.T) {
	tcs := []testCase{
		{"P1", "+O", nil, -1},
		{"P2", "+OK\r", nil, -1},
		{"P3", ":123", nil, -1},
		{"P4", "-Invalid v", nil, -1},
		{"P5", "*3\r\n$4\r\nGood\r\n$7\r\nMorni", Array{}, -1},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, length := ParseFrame([]byte(tc.input))
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Fatalf("Unexpected return value. Expected: %v, Actual: %v", tc.expected, actual)
			}
			if length != tc.length {
				t.Fatalf("Shouldn't have read a full value. Expected: %d, actual: %d", tc.length, length)
			}
		})
	}
}

func TestFull(t *testing.T) {
	tcs := []testCase{
		{"P1", "+OK\r\n", SimpleString{"OK"}, 5},
		{"P2", "-Error parsing\r\n", Error{"Error parsing"}, 16},
		{"P3", ":123\r\n", Integer{123}, 6},
		{"P4", "$4\r\nGood\r\n", BulkString{[]byte("Good")}, 10},
		{"P5", "*3\r\n$4\r\nGood\r\n$7\r\nMorning\r\n$5\r\nFolks\r\n", Array{
			elements: []DataType{
				BulkString{[]byte("Good")},
				BulkString{[]byte("Morning")},
				BulkString{[]byte("Folks")},
			},
		}, 38},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, length := ParseFrame([]byte(tc.input))
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Fatalf("Unexpected return value. Expected: %v, Actual: %v", tc.expected, actual)
			}
			if length != tc.length {
				t.Fatalf("Shouldn't have read a full value. Expected: %d, actual: %d", tc.length, length)
			}
		})
	}
}

func TestFullThenPartial(t *testing.T) {
	tcs := []testCase{
		{"P1", "+OK\r\n$12\r\nSome", SimpleString{"OK"}, 5},
		{"P2", "-Error parsing\r\n+Anothe", Error{"Error parsing"}, 16},
		{"P3", ":123\r\n*3\r\n$2\r\nDi\r\n", Integer{123}, 6},
		{"P4", "$4\r\nGood\r\n+Ano", BulkString{[]byte("Good")}, 10},
		{"P5", "*3\r\n$4\r\nGood\r\n$7\r\nMorning\r\n$5\r\nFolks\r\n$5\r\nGett", Array{
			elements: []DataType{
				BulkString{[]byte("Good")},
				BulkString{[]byte("Morning")},
				BulkString{[]byte("Folks")},
			},
		}, 38},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, length := ParseFrame([]byte(tc.input))
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Fatalf("unexpected return value. Expected: %v, Actual: %v", tc.expected, actual)
			}
			if length != tc.length {
				t.Fatalf("Shouldn't have read a full value. Expected: %d, actual: %d", tc.length, length)
			}
		})
	}
}
