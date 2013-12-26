package message

import (
	"log"
	"math/rand"
	"strconv"
	"testing"

	"code.google.com/p/gogoprotobuf/proto"
	"github.com/go-epaxos/message/example"
)

var start = 0

// Benchmark the Sender / Receiver without serialization (to be compared)
func BenchmardNoSerializetion(b *testing.B) {
	done := make(chan bool, 10)
	if start != 1 {
		startServer(&start)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go startClient(start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

// Benchmark the Sender / Receiver with serialization
func BenchmardWithSerializetion(b *testing.B) {
	done := make(chan bool, 10)
	if start != 2 {
		startServer(&start)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go startClient(start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

// These two functions are copied from examplepb_test.go
// with a change to the size of Deps[] and CommittedDeps[]
func newPreAccept() *example.PreAccept {
	this := &example.PreAccept{}
	this.LeaderId = rand.Int31()
	this.Replica = rand.Int31()
	this.Instance = rand.Int31()
	this.Ballot = rand.Int31()
	v5 := rand.Intn(100)
	this.Command = make([]byte, v5)
	for i := 0; i < v5; i++ {
		this.Command[i] = byte(rand.Intn(256))
	}
	this.Seq = rand.Int31()
	v6 := 5
	this.Deps = make([]int32, v6)
	for i := 0; i < v6; i++ {
		this.Deps[i] = rand.Int31()
	}
	return this
}

func newPreAcceptReply() *example.PreAcceptReply {
	this := &example.PreAcceptReply{}
	this.Replica = rand.Int31()
	this.Instance = rand.Int31()
	this.OK = rand.Uint32()
	this.Ballot = rand.Int31()
	this.Seq = rand.Int31()
	v7 := 5
	this.Deps = make([]int32, v7)
	for i := 0; i < v7; i++ {
		this.Deps[i] = rand.Int31()
	}
	v8 := 5
	this.CommittedDeps = make([]int32, v8)
	for i := 0; i < v8; i++ {
		this.CommittedDeps[i] = rand.Int31()
	}

	return this
}

func startServer(start *int) {
	*start = *start + 1
	r := NewReceiver("localhost:" + strconv.Itoa(8000+*start))
	r.Start()

	// no serialization, just echo
	if *start == 1 {
		go func() {
			for {
				msg := r.Recv()
				msg.reply <- NewMessage(0, msg.Bytes())
			}
		}()
		return
	}
	// with serialization
	if *start == 2 {
		// the server's main loop
		go func() {
			for {
				reply := newPreAcceptReply() // create a reply protobuf
				rmsg := NewEmptyMessage()    // create a reply message
				pa := new(example.PreAccept)
				msg := r.Recv()
				rmsg.msgType = msg.msgType
				// Unmarshal the message bytes
				if err := proto.Unmarshal(msg.Bytes(), pa); err != nil {
					log.Fatal(err)
				}
				// Marshal the reply message bytes
				var err error
				rmsg.bytes, err = proto.Marshal(reply)
				if err != nil {
					log.Fatal(err)
				}
				msg.reply <- rmsg
			}
		}()
		return
	}
}

func startClient(start int, done chan bool) {
	s, err := NewSender("localhost:" + strconv.Itoa(8000+start))
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		// no serialization
		if start == 1 {
			msg := NewMessage(MsgRequireReply+1, []byte("hello"))
			reply, err := s.Send(msg)
			if err != nil {
				log.Fatal(err)
			}
			if string(reply.Bytes()) != "hello" {
				log.Fatal(err)
			}
			continue
		}
		// with serialization
		if start == 2 {
			msg := NewMessage(MsgRequireReply+1, nil)
			pa := newPreAccept() // create a protobuf struct
			pr := new(example.PreAcceptReply)
			// Marshal to bytes
			msg.bytes, err = proto.Marshal(pa)
			if err != nil {
				log.Fatal(err)
			}
			reply, err := s.Send(msg)
			if err != nil {
				log.Fatal(err)
			}
			//Unmarshl the reply bytes
			if err := proto.Unmarshal(reply.Bytes(), pr); err != nil {
				log.Fatal(err)
			}
			continue
		}
	}
	done <- true
}
