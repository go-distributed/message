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
	buf, msg := initMsg(t)

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
	compareMsg(msg, outMsg, t)
}

func TestNonBlockRecv(t *testing.T) {
	buf, msg := initMsg(t)

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
	compareMsg(msg, outMsg, t)
}

func TestPbBlockRecv(t *testing.T) {
	register(0, reflect.TypeOf(example.PreAccept{}))
	sp := NewPreAcceptSample()
	r := NewPbReceiver(":8000")
	r.GoStart()
	defer r.Stop()

	c := make(chan *PbMessage)
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
	sender, err := NewPbSender(":8000")
	if err != nil {
		t.Fatal(err)
	}
	inMsg := NewPbMessage(0, sp)
	go sender.Send(inMsg)
	outMsg := <-c
	compareMsg(inMsg, outMsg, t)
}

func TestPbNonBlockRecv(t *testing.T) {
	register(0, reflect.TypeOf(example.PreAccept{}))
	sp := NewPreAcceptSample()
	r := NewPbReceiver(":8001")
	r.GoStart()
	defer r.Stop()

	time.Sleep(50 * time.Millisecond) // wait for server to start

	// should not receive anything
	outMsg := r.GoRecv()
	if outMsg != nil {
		t.Fatal("GoRecv() return not nil")
	}

	// start sending
	sender, err := NewPbSender(":8001")
	if err != nil {
		t.Fatal(err)
	}
	inMsg := NewPbMessage(0, sp)
	go sender.Send(inMsg)
	time.Sleep(50 * time.Millisecond) // wait for socket to be available

	outMsg = r.GoRecv()
	if outMsg == nil {
		t.Fatal("nil")
	}
	compareMsg(inMsg, outMsg, t)
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

func TestSendPbNil(t *testing.T) {
	register(0, reflect.TypeOf(example.PreAccept{}))
	r := NewPbReceiver(":8002")
	r.GoStart()
	defer r.Stop()
	time.Sleep(50 * time.Millisecond)

	sender, err := NewPbSender(":8002")
	if err != nil {
		t.Fatal(err)
	}
	inMsg := NewPbMessage(0, nil)
	go sender.Send(inMsg)
	time.Sleep(50 * time.Millisecond) // wait for GoRecv

	out := r.GoRecv()
	compareMsg(out, inMsg, t)
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

// Test whether receiver can stop and listen again for PbReceiver
func TestPbStopRestart(t *testing.T) {
	r := NewPbReceiver(":8003")
	r.GoStart()
	defer r.Stop()
	time.Sleep(50 * time.Millisecond)

	r.Stop()
	r.GoStart()
	time.Sleep(50 * time.Millisecond) // prevent running r.Stop() before r.GoStart()
}

// Test multiple stop for PbReceiver
func TestPbMultipleStop(t *testing.T) {
	r := NewPbReceiver(":8004")
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
	buf, msg := initMsg(t)
	secBuf, _ := initMsg(t)
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
			compareMsg(msg, outMsg, t)
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

// Test multiple pbmessage
func TestMultiplePbMessage(t *testing.T) {
	register(0, reflect.TypeOf(example.PreAccept{}))
	cnt := 3 // number of messages
	finish := make(chan bool)

	// start receiver
	r := NewPbReceiver(":8005")
	r.GoStart()
	defer r.Stop()
	time.Sleep(50 * time.Millisecond)

	// start sending
	sender, err := NewPbSender(":8005")
	if err != nil {
		t.Fatal(err)
	}
	sp := NewPreAcceptSample()
	inMsg := NewPbMessage(0, sp)
	for i := 0; i < cnt; i++ {
		sender.Send(inMsg)
	}
	// test correctness
	go func() {
		for i := 0; i < cnt; i++ {
			outMsg := r.Recv()
			compareMsg(inMsg, outMsg, t)
		}
		finish <- true
	}()
	// test timeout
	for {
		select {
		case <-finish:
			return
		case <-time.After(5 * time.Second):
			t.Fatal("Did not receive two pbmessages!")
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
	defer r.Stop()

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

// Test send out a pbmessage to a local receiver and wait for reply
func TestPbSendTo(t *testing.T) {
	register(MsgRequireReply+1, reflect.TypeOf(example.PreAccept{}))
	r := NewPbReceiver(":8007")
	r.GoStart()
	defer r.Stop()

	sp := NewPreAcceptSample()
	m := NewPbMessage(MsgRequireReply+1, sp)
	go func() {
		msg := r.Recv()
		msg.reply <- msg
	}()

	reply := PbSendTo(r, m)

	compareMsg(m, reply, t)
}

func initMsg(t *testing.T) (*bytes.Buffer, *Message) {
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

	return buf, msg
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

func compareMsg(msg, outMsg interface{}, t *testing.T) {
	if !reflect.DeepEqual(msg, outMsg) {
		t.Fatal("Messages are not equal!")
	}

}

func mockServer(addr string) {
	r := NewReceiver(addr)
	r.GoStart()

	msg := r.Recv()
	msg.reply <- NewMessage(0, append([]byte("a reply to "), msg.bytes...))
}

func mockPbServer(addr string) {
	register(MsgRequireReply+1, reflect.TypeOf(example.A{}))
	r := NewPbReceiver(addr)
	r.GoStart()

	msg := r.Recv()
	msg.reply <- NewPbMessage(msg.Type(), msg.Proto())
}
