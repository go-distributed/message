package message

import (
	"testing"
	"time"
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

	if string(reply.Bytes()) != "a reply" {
		t.Fatal("error recv!")
	}
}
