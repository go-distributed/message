package message

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-epaxos/message/example"
)

func TestSend(t *testing.T) {
	addr := ":9000"
	go mockServer(addr)

	// wait for the server to be available
	time.Sleep(50 * time.Millisecond)

	sender, err := NewSender(addr)
	if err != nil {
		t.Fatal(err)
	}

	// send out a message which need reply
	msg := NewMessage(MsgRequireReply+1, []byte("a send"))
	reply, err := sender.Send(msg)
	if err != nil {
		t.Fatal(err)
	}

	if string(reply.Bytes()) != "a reply to a send" {
		t.Fatal("error recv!")
	}
}

func TestSendPb(t *testing.T) {
	addr := ":9001"
	go mockPbServer(addr)

	// wait for the server to be available
	time.Sleep(50 * time.Millisecond)

	sender, err := NewPbSender(addr)
	if err != nil {
		t.Fatal(err)
	}

	// send out a message which need reply
	ex := &example.A{
		Description: "hello",
		Number:      42,
	}
	msg := NewPbMessage(MsgRequireReply+1, ex)
	reply, err := sender.Send(msg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(ex, reply.pb) {
		t.Fatal("error recv!, result not equal")
	}
}
