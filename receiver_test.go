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

// TestBlockRecv tests blocking receive.
// The Recv() function should block until receive a message from a sender.
func TestBlockRecv(t *testing.T) {
	buf, inPb, msg := initMsg(t)

	r := NewReceiver(":8000")
	r.GoStart()
	defer r.Stop()

	c := make(chan *Message)
	go func() {
		c <- r.Recv()
	}()

	// before the sender sends the message, Recv() should not return.
	select {
	case <-time.After(50 * time.Millisecond):
	case <-c:
		t.Fatal("Recv should block until sender sends message")
	}

	// start sending
	go send("8000", buf, t)
	outMsg := <-c
	compareMsg(msg, outMsg, inPb, t)
}

func TestNonBlockRecv(t *testing.T) {
	buf, inPb, msg := initMsg(t)

	// start receiver
	r := NewReceiver(":8001")
	r.GoStart()
	defer r.Stop()

	time.Sleep(50 * time.Millisecond) // wait for server to start

	// should not receive anything
	outMsg := r.GoRecv()
	if outMsg != nil {
		t.Fatal("GoRecv() return not nil")
	}

	// start sending
	go send("8001", buf, t)
	time.Sleep(50 * time.Millisecond) // wait for socket to be available

	outMsg = r.GoRecv()
	if outMsg == nil {
		t.Fatal("nil")
	}
	compareMsg(msg, outMsg, inPb, t)
}

func TestSendTrash(t *testing.T) {
	r := NewReceiver(":8002")
	r.GoStart()
	defer r.Stop()
	time.Sleep(50 * time.Millisecond)

	conn, err := net.Dial("tcp", ":8002")
	if err != nil {
		t.Fatal(err)
	}

	n, err := bufio.NewWriter(conn).WriteString("Some evil trash hahaha")
	if err != nil {
		t.Log(n)
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond) // wait for GoRecv
	if r.GoRecv() != nil {
		t.Fatal("Should not receive anything!")
	}
}

// Test whether receiver can stop and listen again
func TestStopRestart(t *testing.T) {
	r := NewReceiver(":8003")
	r.GoStart()
	defer r.Stop()
	time.Sleep(50 * time.Millisecond)

	r.Stop()
	r.GoStart()
	time.Sleep(50 * time.Millisecond) // prevent running r.Stop() before r.GoStart()
}

// Test multiple stop
func TestMultipleStop(t *testing.T) {
	r := NewReceiver(":8004")
	r.GoStart()
	defer r.Stop()
	time.Sleep(50 * time.Millisecond)

	for i := 0; i < 5; i++ {
		r.Stop()
	}
	r.GoStart()
	time.Sleep(50 * time.Millisecond) // prevent running r.Stop() before r.GoStart()
}

// Test multiple message
func TestMultipleMessage(t *testing.T) {
	finish := make(chan bool)
	buf, inPb, msg := initMsg(t)
	secBuf, _, _ := initMsg(t)
	secBuf.WriteTo(buf) // write second message

	r := NewReceiver(":8005")
	r.GoStart()
	defer r.Stop()
	time.Sleep(50 * time.Millisecond)

	// start sending
	go send("8005", buf, t)

	// test correctness
	go func() {
		for i := 0; i < 2; i++ {
			outMsg := r.Recv()
			compareMsg(msg, outMsg, inPb, t)
		}
		finish <- true
	}()
	// test timeout
	for {
		select {
		case <-finish:
			return
		case <-time.After(5 * time.Second):
			t.Fatal("Did not receive two messages!")
		}
	}
}

// Test send out a message and wait for reply
func TestSendAndReply(t *testing.T) {
	addr := ":8006"

	go mockServer(addr)

	// wait for server to start
	time.Sleep(time.Millisecond * 50)
	sendAndRecv(addr, t)
}

// Test send out a message to a local receiver and wait for reply
func TestSendTo(t *testing.T) {
	r := NewReceiver(":8007")
	r.GoStart()

	go func() {
		msg := r.Recv()
		msg.reply <- NewMessage(0, append([]byte("a reply to "), msg.bytes...))
	}()

	m := NewMessage(MsgRequireReply+1, []byte("a send"))
	reply := SendTo(r, m)

	if string(reply.Bytes()) != "a reply to a send" {
		t.Fatal("recv error")
	}
}

func initMsg(t *testing.T) (*bytes.Buffer, *example.A, *Message) {
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

func sendAndRecv(addr string, t *testing.T) {
	// Dial
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	// send out a message which need reply
	msg := NewMessage(MsgRequireReply+1, []byte("a send"))
	e := NewMsgEncoder(conn)
	err = e.Encode(msg)
	if err != nil {
		t.Fatal(err)
	}

	// receive reply
	d := NewMsgDecoder(conn)
	reply := NewEmptyMessage()
	d.Decode(reply)

	if string(reply.Bytes()) != "a reply to a send" {
		t.Fatal("error recv!")
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

func mockServer(addr string) {
	r := NewReceiver(addr)
	r.GoStart()

	msg := r.Recv()
	msg.reply <- NewMessage(0, append([]byte("a reply to "), msg.bytes...))
}

func mockPbServer(addr string) {
	r := NewPbReceiver(addr)
	register(MsgRequireReply+1, reflect.TypeOf(example.A{}))
	r.GoStart()

	msg := r.Recv()
	msg.reply <- NewPbMessage(msg.Type(), msg.Proto())
}










