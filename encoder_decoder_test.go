package message

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"code.google.com/p/gogoprotobuf/proto"
	"github.com/xiangli-cmu/message/example"
)

// a simple test that encodes a message into a buffer
// and decodes a message out of the buffer.
func TestEncoderAndDecoder(t *testing.T) {
	buf := new(bytes.Buffer)

	inPb := &example.A{
		Description: "hello world!",
		Number:      1,
		Id:          []byte{0x00},
	}

	bytes, err := proto.Marshal(inPb)

	if err != nil {
		t.Fatal(err)
	}

	msg := NewMessage(0, bytes)

	e := NewMsgEncoder(buf)
	e.Encode(msg)

	outMsg := NewEmptyMessage()

	d := NewMsgDecoder(buf)
	d.Decode(outMsg)

	if !reflect.DeepEqual(msg, outMsg) {
		t.Fatal("Messages are not equal!")
	}

	outPb := new(example.A)

	proto.Unmarshal(outMsg.Bytes(), outPb)

	if !reflect.DeepEqual(outPb, inPb) {
		fmt.Println(outPb.GetId())
		fmt.Println(inPb.GetId())
		t.Fatal("Protos are not equal!")
	}
}
