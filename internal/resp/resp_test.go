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
package resp

import (
	"reflect"
	"testing"
)

type testCase struct {
	name     string
	input    string
	expected any
	err      error
}

func TestPartial(t *testing.T) {
	tcs := []testCase{
		{"P1", "+O", nil, nil},
		{"P2", "+OK\r", nil, nil},
		{"P3", ":123", nil, nil},
		{"P4", "-Invalid v", nil, nil},
		{"P5", "*3\r\n$4\r\nGood\r\n$7\r\nMorni", Array{}, nil},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ParseFrame([]byte(tc.input))
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Fatalf("Unexpected return value. Expected: %v, Actual: %v", tc.expected, actual)
			}
			if err != tc.err {
				t.Fatalf("Unexpected error value. Expected: %v, Actual: %v", tc.err, err)
			}
		})
	}
}

func TestFull(t *testing.T) {
	tcs := []testCase{
		{"P1", "+OK\r\n", SimpleString{"OK"}, nil},
		{"P2", "-Error parsing\r\n", SimpleError{"Error parsing"}, nil},
		{"P3", ":123\r\n", IntegerValue{123}, nil},
		{"P4", "$4\r\nGood\r\n", BulkString{[]byte("Good")}, nil},
		{"P5", "*3\r\n$4\r\nGood\r\n$7\r\nMorning\r\n$5\r\nFolks\r\n", Array{
			values: []DataType{
				BulkString{[]byte("Good")},
				BulkString{[]byte("Morning")},
				BulkString{[]byte("Folks")},
			},
		}, nil},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ParseFrame([]byte(tc.input))
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Fatalf("Unexpected return value. Expected: %v, Actual: %v", tc.expected, actual)
			}
			if err != tc.err {
				t.Fatalf("Unexpected error value. Expected: %v, Actual: %v", tc.err, err)
			}
		})
	}
}

func TestFullThenPartial(t *testing.T) {
	tcs := []testCase{
		{"P1", "+OK\r\n$12\r\nSome", SimpleString{"OK"}, nil},
		{"P2", "-Error parsing\r\n+Anothe", SimpleError{"Error parsing"}, nil},
		{"P3", ":123\r\n*3\r\n$2\r\nDi\r\n", IntegerValue{123}, nil},
		{"P4", "$4\r\nGood\r\n+Ano", BulkString{[]byte("Good")}, nil},
		{"P5", "*3\r\n$4\r\nGood\r\n$7\r\nMorning\r\n$5\r\nFolks\r\n$5\r\nGett", Array{
			values: []DataType{
				BulkString{[]byte("Good")},
				BulkString{[]byte("Morning")},
				BulkString{[]byte("Folks")},
			},
		}, nil},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ParseFrame([]byte(tc.input))
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Fatalf("Unexpected return value. Expected: %v, Actual: %v", tc.expected, actual)
			}
			if err != tc.err {
				t.Fatalf("Unexpected error value. Expected: %v, Actual: %v", tc.err, err)
			}
		})
	}
}
