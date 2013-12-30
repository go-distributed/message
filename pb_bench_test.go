package message

import (
	"log"
	"reflect"
	"testing"

	"github.com/go-epaxos/message/example"
)

var (
	// To control the server from starting multiple times
	pbNoSerializationServerStarted   = false
	pbWithSerializationServerStarted = false
)

// Benchmark the Sender / Receiver without serialization (to be compared)
func BenchmarkPbNoSerializetion(b *testing.B) {
	initTest()
	done := make(chan bool, 10)
	if !pbNoSerializationServerStarted {
		startPbServerNoSerialization()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go startPbClientNoSerialization(done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

// Benchmark the PbSender / Receiver with serialization
func BenchmarkPbWithSerializetion(b *testing.B) {
	initTest()
	done := make(chan bool, 10)
	if !pbWithSerializationServerStarted {
		startPbServerWithSerialization()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go startPbClientWithSerialization(done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

func initTest() {
	register(MsgRequireReply+1, reflect.TypeOf(example.PreAccept{}))
	register(MsgRequireReply+2, reflect.TypeOf(example.PreAcceptReply{}))
}

func startPbServerNoSerialization() {
	pbNoSerializationServerStarted = true
	r := NewPbReceiver("localhost:8002")
	r.GoStart()

	// no serialization, just echo
	go func() {
		for {
			msg := r.Recv()
			msg.reply <- NewPbMessage(msg.Type()+1, msg.Proto())
		}
	}()
}

func startPbServerWithSerialization() {
	pbWithSerializationServerStarted = true
	r := NewPbReceiver("localhost:8003")
	r.GoStart()

	// with serialization
	go func() {
		for {
			msg := r.Recv()
			rmsg := NewPbMessage(msg.Type()+1, NewPreAcceptReplySample()) // the reply PbMessage
			msg.reply <- rmsg
		}
	}()
}

func startPbClientNoSerialization(done chan bool) {
	s, err := NewPbSender("localhost:8002")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		msg := NewPbMessage(MsgRequireReply+1, nil)
		reply, err := s.Send(msg)
		if err != nil {
			log.Fatal(err)
		}
		if reply.Type() != (msg.Type() + 1) {
			log.Fatal("Unexpected reply msgType:", reply.Type())
		}
	}
	done <- true
}

func startPbClientWithSerialization(done chan bool) {
	// with serialization
	s, err := NewPbSender("localhost:8003")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		msg := NewPbMessage(MsgRequireReply+1, NewPreAcceptSample())
		reply, err := s.Send(msg)
		if err != nil {
			log.Fatal(err)
		}
		if reply.Type() != (msg.Type() + 1) {
			log.Fatal("Unexpected reply msgType:", reply.Type())
		}
	}
	done <- true
}
