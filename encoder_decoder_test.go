package message

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/xiangli-cmu/message/example"
)

// a simple test that encodes a pb message into a buffer
// and decode a pb message out of the buffer.
// If the two pb messages are equal, the test will be passed.
func TestEncoderAndDecoder(t *testing.T) {
	buf := new(bytes.Buffer)

	inPb := &example.A{
		Description: "hello world!",
		Number:      1,
		Id:          []byte{0x00, 0x01, 0x02, 0x03},
	}

	outPb := new(example.A)

	e := NewPbEecoder(buf)
	e.Encode(inPb)

	d := NewPbDecoder(buf)
	d.Decode(outPb)

	if reflect.DeepEqual(outPb, inPb) {
		t.Fatal("Not equal!")
	}
}
