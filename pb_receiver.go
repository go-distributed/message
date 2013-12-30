package message

import (
	"io"
	"net"
	"time"

	"github.com/coreos/go-log/log"
)

// Receiver struct
type PbReceiver struct {
	localAddr    *net.TCPAddr     // address
	ln           *net.TCPListener // only TCP now
	ch           chan *PbMessage  // message channel
	stop         bool             // stop?
	replyTimeout time.Duration
}

// Constructor
func NewPbReceiver(addrStr string) *PbReceiver {
	r := new(PbReceiver)
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		log.Error("ResolveTCPAddr() error: ", err)
		return nil
	}
	r.localAddr = addr
	r.ch = make(chan *PbMessage, chanBufSize)
	// TODO: this should be configurable
	r.replyTimeout = time.Millisecond * 50
	return r
}

// Recv() will blocking until there is message
func (r *PbReceiver) Recv() *PbMessage {
	return <-r.ch
}

// GoRecv() will return message if possible, or nil if no message
func (r *PbReceiver) GoRecv() *PbMessage {
	select {
	case m := <-r.ch:
		return m
	default:
		return nil
	}
}

// Send a message to a local receiver
func PbSendTo(r *PbReceiver, m *PbMessage) *PbMessage {
	attached := m.AttachReplyChan()
	r.ch <- m
	if attached {
		reply := <-m.reply
		return reply
	}
	return nil
}

func (r *PbReceiver) GoStart() {
	go r.Start()
}

// Stop the receiver
func (r *PbReceiver) Stop() error {
	r.stop = true
	err := r.ln.Close()
	if err != nil {
		return err
	}
	return nil
}

// Start listen and receive messages
func (r *PbReceiver) Start() {
	ln, err := net.ListenTCP("tcp", r.localAddr)
	if err != nil {
		log.Error("Listen() error: ", err)
		return
	}
	r.ln = ln
	for {
		conn, err := r.ln.AcceptTCP()
		if err != nil {
			if r.stop {
				return
			}
			log.Warning("Accept() error: ", err)
			// TODO: is it a temp error?
			// need to check!
			continue
		}
		go r.handleConn(conn)
	}
}

// handleConn handles incoming connections
// It decodes a message from TCP stream and sends it to channel
func (r *PbReceiver) handleConn(conn net.Conn) {
	d := NewMsgDecoder(conn)
	e := NewMsgEncoder(conn)

	for {
		// create an empty message with reply channel
		msg := NewEmptyPbMessage()

		err := d.DecodePb(msg)
		if err != nil {
			if err == io.EOF {
				return
			}
			// TODO: handle error
			log.Warning("handleConn() error: ", err)
			return
		}

		attached := msg.AttachReplyChan()

		// send received message for processing
		r.ch <- msg

		if attached {
			// wait for reply
			replyMsg := <-msg.reply
			if replyMsg != nil {
				if err := e.EncodePb(replyMsg); err != nil {
					if err == io.EOF {
						return
					}
					// TODO: handle error
					log.Warning("handleConn() error: ", err)
				}
			}
		}
	}
}
