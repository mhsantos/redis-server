package commands

import (
	"fmt"
	"testing"

	"github.com/mhsantos/redis-server/internal/protocol"
)

type delTestCase struct {
	name      string
	input     string
	setupCmds []protocol.Array
	expected  any
}

func TestDelInvalid(t *testing.T) {
	dtcs := []delTestCase{
		{
			name:     "Invalid number",
			input:    "*1\r\n$3\r\nDEL\r\n",
			expected: protocol.NewError(delInvalidLengthErrMsg),
		},
		{
			name:     "Invalid key type",
			input:    "*2\r\n$3\r\nDEL\r\n:1\r\n",
			expected: protocol.NewError(fmt.Sprintf(delKeyTypeErrMsg, protocol.NewInteger(1))),
		},
	}
	for _, tc := range dtcs {
		t.Run(tc.name, func(t *testing.T) {
			arguments, _ := ParseCommand([]byte(tc.input))
			actual := ProcessCommand(arguments.(protocol.Array))
			if actual != tc.expected {
				t.Fatalf("unexpected return value. Expected: %v, Actual: %v", tc.expected, actual)
			}
		})
	}

}

func TestDelValid(t *testing.T) {
	dtcs := []delTestCase{
		{
			name:  "3 keys with 2 existing",
			input: "*4\r\n$3\r\nDEL\r\n$4\r\nname\r\n$8\r\nlastname\r\n$3\r\nage\r\n",
			setupCmds: []protocol.Array{
				protocol.NewArray(
					protocol.NewBulkString([]byte("set")),
					protocol.NewBulkString([]byte("name")),
					protocol.NewBulkString([]byte("john")),
				),
				protocol.NewArray(
					protocol.NewBulkString([]byte("set")),
					protocol.NewBulkString([]byte("age")),
					protocol.NewInteger(20),
				),
			},
			expected: protocol.NewInteger(2),
		},
		{
			name:  "4 keys with 2 existing",
			input: "*5\r\n$3\r\nDEL\r\n$5\r\nmonth\r\n$3\r\nday\r\n$4\r\nyear\r\n$4\r\nhour\r\n",
			setupCmds: []protocol.Array{
				protocol.NewArray(
					protocol.NewBulkString([]byte("set")),
					protocol.NewBulkString([]byte("month")),
					protocol.NewBulkString([]byte("May")),
				),
				protocol.NewArray(
					protocol.NewBulkString([]byte("set")),
					protocol.NewBulkString([]byte("day")),
					protocol.NewInteger(15),
				),
			},
			expected: protocol.NewInteger(2),
		},
	}
	for _, tc := range dtcs {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.setupCmds {
				ProcessCommand(cmd)
			}
			arguments, _ := ParseCommand([]byte(tc.input))
			actual := ProcessCommand(arguments.(protocol.Array))
			if actual != tc.expected {
				t.Fatalf("unexpected return value. Expected: %v, Actual: %v", tc.expected, actual)
			}
		})
	}
}
