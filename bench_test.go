package message

import (
	"log"
	"math/rand"
	"testing"

	"code.google.com/p/gogoprotobuf/proto"
	"github.com/go-epaxos/message/example"
)

var (
	// To control the server from starting multiple times
	NoSerializationServerStarted   = false
	WithSerializationServerStarted = false
)

// Benchmark the Sender / Receiver without serialization (to be compared)
func BenchmarkNoSerializetion(b *testing.B) {
	done := make(chan bool, 10)
	if !NoSerializationServerStarted {
		startServerNoSerialization()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go startClientNoSerialization(done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

// Benchmark the Sender / Receiver with serialization
func BenchmarkWithSerializetion(b *testing.B) {
	done := make(chan bool, 10)
	if !WithSerializationServerStarted {
		startServerWithSerialization()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go startClientWithSerialization(done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

// These two functions are modified from example.pb.go
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

func startServerNoSerialization() {
	NoSerializationServerStarted = true
	r := NewReceiver("localhost:8000")
	r.GoStart()

	// no serialization, just echo
	go func() {
		for {
			msg := r.Recv()
			msg.reply <- NewMessage(0, msg.Bytes())
		}
	}()
}

func startServerWithSerialization() {
	WithSerializationServerStarted = true
	r := NewReceiver("localhost:8001")
	r.GoStart()

	// with serialization
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
}

func startClientNoSerialization(done chan bool) {
	s, err := NewSender("localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		msg := NewMessage(MsgRequireReply+1, []byte("hello"))
		reply, err := s.Send(msg)
		if err != nil {
			log.Fatal(err)
		}
		if string(reply.Bytes()) != "hello" {
			log.Fatal(err)
		}
	}
	done <- true
}

func startClientWithSerialization(done chan bool) {
	// with serialization
	s, err := NewSender("localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
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
	}
	done <- true
}
