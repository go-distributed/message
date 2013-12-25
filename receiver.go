package message

import (
	"github.com/coreos/go-log/log"
	"io"
	"net"
)

const (
	chanBufSize = 10 // buffer size for message channel, default = 10
)

// Receiver struct
type Receiver struct {
	addr string           // address, in "[ip]:port" format
	ch   chan *Message    // message channel
	ln   *net.TCPListener // only TCP now
	stop bool             // stop?
}

// Constructor
func NewReceiver(addr string) *Receiver {
	r := new(Receiver)
	r.addr = addr
	r.ch = make(chan *Message, chanBufSize)
	return r
}

// Recv() will blocking until there is message
func (r *Receiver) Recv() *Message {
	return <-r.ch
}

// GoRecv() will return message if possible, or nil if no message
func (r *Receiver) GoRecv() *Message {
	select {
	case m := <-r.ch:
		return m
	default:
		return nil
	}
}

func (r *Receiver) Start() {
	go r.start()
}

// Stop the receiver
func (r *Receiver) Stop() error {
	r.stop = true
	err := r.ln.Close()
	if err != nil {
		return err
	}
	return nil
}

// start listen and receive messages
func (r *Receiver) start() {
	addr, err := net.ResolveTCPAddr("tcp", r.addr)
	if err != nil {
		log.Error("ResolveTCPAddr() error:", err)
		return
	}

	r.ln, err = net.ListenTCP("tcp", addr)
	if err != nil {
		log.Error("Listen() error:", err)
		return
	}

	for {
		conn, err := r.ln.AcceptTCP()
		if err != nil {
			if r.stop {
				return
			}
			log.Warning("Accept() error:", err)
			// TODO: is it a temp error?
			// need to check!
			continue
		}
		go r.handleConn(conn)
	}
}

// handleConn handles incoming connections
// It decodes a message from TCP stream and sends it to channel
func (r *Receiver) handleConn(conn net.Conn) {
	d := NewMsgDecoder(conn)

	for {
		msg := NewEmptyMessage()
		err := d.Decode(msg)
		if err != nil {
			if err != io.EOF {
				log.Warning("Decode() error:", err)
			}
			return
		}
		r.ch <- msg
	}
}
