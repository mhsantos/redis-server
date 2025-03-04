package commands

import (
	"testing"

	"github.com/mhsantos/redis-server/internal/protocol"
)

type commandTestCase struct {
	name     string
	input    string
	expected any
	length   int
}

func TestCommandParsing(t *testing.T) {
	ctc := []commandTestCase{
		{"Invalid command", "*3\r\n$3\r\nbuy\r\n$3\r\nkey\r\n$3\r\nval\r\n", protocol.NewError("invalid command buy"), 31},
		{"Invalid GET arguments", "*3\r\n$3\r\nGET\r\n$3\r\nkey\r\n$4\r\nabcd\r\n", protocol.NewError("the GET command accepts 2 parameters: GET and KEY. Received 3 parameters instead"), 32},
	}
	for _, tc := range ctc {
		t.Run(tc.name, func(t *testing.T) {
			arguments, length := ParseCommand([]byte(tc.input))
			actual := ProcessCommand(arguments.(protocol.Array))
			if actual != tc.expected {
				t.Fatalf("unexpected return value. Expected: %v, Actual: %v", tc.expected, actual)
			}
			if length != tc.length {
				t.Fatalf("incorrect number of bytes read. Expected %d, actual %d", tc.length, length)
			}
		})
	}
}
