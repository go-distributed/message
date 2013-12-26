package message

import (
	"io"
	"net"
	"time"

	"github.com/coreos/go-log/log"
)

const (
	chanBufSize = 10 // buffer size for message channel, default = 10
)

// Receiver struct
type Receiver struct {
	addr         string           // address, in "[ip]:port" format
	ch           chan *Message    // message channel
	ln           *net.TCPListener // only TCP now
	stop         bool             // stop?
	replyTimeout time.Duration
}

// Constructor
func NewReceiver(addr string) *Receiver {
	r := new(Receiver)
	r.addr = addr
	r.ch = make(chan *Message, chanBufSize)
	// TODO: this should be configurable
	r.replyTimeout = time.Millisecond * 50
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

func (r *Receiver) GoStart() {
	go r.Start()
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

// Start listen and receive messages
func (r *Receiver) Start() {
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
	e := NewMsgEncoder(conn)

	for {
		// create an empty message with reply channel
		msg := NewEmptyMessage()

		err := d.Decode(msg)
		if err != nil {
			if err == io.EOF {
				return
			}
			// TODO: handle error
			log.Warning("handleConn() error:", err)
			return
		}

		attached := msg.AttachReplyChan()

		// send received message for processing
		r.ch <- msg

		if attached {
			// wait for reply
			replyMsg := <-msg.reply
			if replyMsg != nil {
				if err := e.Encode(replyMsg); err != nil {
					if err == io.EOF {
						return
					}
					// TODO: handle error
					log.Warning("handleConn() error:", err)
				}
			}
		}
	}
}
