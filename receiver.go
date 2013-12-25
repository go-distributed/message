package message

import (
	"log"
	"net"
)

const (
	CHANSIZE = 10
)

type Receiver struct {
	port string
	ch   chan *Message
	ln   net.Listener
	stop chan bool
}

func NewReceiver(port string) *Receiver {
	r := new(Receiver)
	r.port = port
	r.ch = make(chan *Message, CHANSIZE)
	r.stop = make(chan bool, 1)
	return r
}

func (r *Receiver) Recv() *Message {
	return <-r.ch
}

func (r *Receiver) GoRecv() *Message {
	select {
	case m := <-r.ch:
		return m
	default:
		return nil
	}
}

func (r *Receiver) Start() {
	ln, err := net.Listen("tcp", ":"+r.port) // only TCP now
	r.ln = ln
	if err != nil {
		log.Fatal("Listen() error:", err)
	}

Loop:
	for {
		select {
		case <-r.stop:
			break Loop
		default:
			conn, err := ln.Accept()
			if err != nil {
				log.Println("Accept() error:", err)
				continue
			}
			go func() {
				msg := NewEmptyMessage()
				d := NewMsgDecoder(conn)
				err := d.Decode(msg)
				if err != nil {
					log.Println("Decode() error:", err)
					return
				}
				r.ch <- msg
			}()
		}
	}
}

func (r *Receiver) Stop() error {
	r.stop <- true
	return r.ln.Close()
}
