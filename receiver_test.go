package message

import (
	"bufio"
	"bytes"
	"net"
	"reflect"
	"testing"
	"time"

	"code.google.com/p/gogoprotobuf/proto"
	"github.com/go-epaxos/message/example"
)

func TestBlockRecv(t *testing.T) {
	buf, inPb, msg := inittest(t)
	tmpChan := make(chan *Message)
	r := NewReceiver("8000")
	go r.Start()
	time.Sleep(2 * time.Second)

	// should not receive anything
	go func() {
		tmpChan <- r.Recv()
	}()

Loop:
	for {
		select {
		case <-time.After(3 * time.Second):
			break Loop
		case <-tmpChan:
			t.Fatal("Recv should block!")
		}
	}

	// start sending
	go send("8000", buf, t)
	outMsg := <-tmpChan
	compareMsg(msg, outMsg, inPb, t)
}

func TestNonBlockRecv(t *testing.T) {
	buf, inPb, msg := inittest(t)

	// start receiver
	r := NewReceiver("8001")
	go r.Start()
	time.Sleep(2 * time.Second) // wait for server to start

	// should not receive anything
	outMsg := r.GoRecv()
	if outMsg != nil {
		t.Fatal("GoRecv() return not nil")
	}

	// start sending
	go send("8001", buf, t)
	time.Sleep(2 * time.Second) // wait for socket to be available

	outMsg = r.GoRecv()
	if outMsg == nil {
		t.Fatal("nil")
	}
	compareMsg(msg, outMsg, inPb, t)
}

func TestSendTrash(t *testing.T) {
	r := NewReceiver("8002")
	go r.Start()
	time.Sleep(2 * time.Second)

	conn, err := net.Dial("tcp", ":8002")
	if err != nil {
		t.Fatal(err)
	}

	n, err := bufio.NewWriter(conn).WriteString("Some evil trash hahaha")
	if err != nil {
		t.Log(n)
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second) // wait for GoRecv
	if r.GoRecv() != nil {
		t.Fatal("Should not receive anything!")
	}
}

// Test whether receiver can stop and listen again
func TestStopRestart(t *testing.T) {
	r := NewReceiver("8003")
	go r.Start()
	time.Sleep(2 * time.Second)

	r.Stop()
	go r.Start()
}

func inittest(t *testing.T) (*bytes.Buffer, *example.A, *Message) {
	buf := new(bytes.Buffer)
	inPb := &example.A{
		Description: "hello world!",
		Number:      1,
	}
	// UUID is 16 byte long
	for i := 0; i < 16; i++ {
		inPb.Id = append(inPb.Id, byte(i))
	}

	bytes, err := proto.Marshal(inPb)
	if err != nil {
		t.Fatal(err)
	}

	msg := NewMessage(0, bytes)
	e := NewMsgEncoder(buf)
	e.Encode(msg)

	return buf, inPb, msg
}

func send(port string, buf *bytes.Buffer, t *testing.T) {
	conn, err := net.Dial("tcp", ":"+port)
	if err != nil {
		t.Fatal(err)
	}

	n, err := buf.WriteTo(conn)
	if err != nil {
		t.Log(n)
		t.Fatal(err)
	}
}

func compareMsg(msg, outMsg *Message, a interface{}, t *testing.T) {
	if !reflect.DeepEqual(msg, outMsg) {
		t.Fatal("Messages are not equal!")
	}

	switch reflect.TypeOf(a).String() {
	case "*example.A":
		outPb := new(example.A)
		proto.Unmarshal(outMsg.bytes, outPb)
		if !reflect.DeepEqual(outPb, outPb) {
			t.Fatal("Protos are not equal!")
		}
	default:
		t.Fatal("Unknow type")
	}
}
