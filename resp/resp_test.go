/*
# Redis Protocol:
# Spec: https://redis.io/docs/latest/develop/reference/protocol-spec/

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

import "testing"

func TestPartial(t *testing.T) {
	test := []byte("+partial")
	expected := SimpleString{}
	actual, err := ParseFrame(test)
	if err != nil {
		t.Fatal("Shouldn't error")
	}
	if expected != actual {
		t.Fatal("Expected different than actual")
	}
}

func TestFull(t *testing.T) {
	test := []byte("+something\r\n")
	expected := SimpleString{
		data: "something",
	}
	actual, err := ParseFrame(test)
	if err != nil {
		t.Fatal("Shouldn't error")
	}
	if expected != actual {
		t.Fatal("Expected different than actual")
	}
}

func TestFullThenPartial(t *testing.T) {
	test := []byte("+something\r\nelse")
	expected := SimpleString{
		data: "something",
	}
	actual, err := ParseFrame(test)
	if err != nil {
		t.Fatal("Shouldn't error")
	}
	if expected != actual {
		t.Fatal("Expected different than actual")
	}
}
