package message

import (
	"net"
)

type Sender struct {
	remoteAddr *net.TCPAddr
	conn       *net.TCPConn
	encoder    *MsgEncoder
	decoder    *MsgDecoder
}

func NewSender(raddrStr string) (*Sender, error) {
	raddr, err := net.ResolveTCPAddr("tcp", raddrStr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}

	return &Sender{
		remoteAddr: raddr,
		conn:       conn,
		encoder:    NewMsgEncoder(conn),
		decoder:    NewMsgDecoder(conn),
	}, nil
}

func (s *Sender) Send(msg *Message) (*Message, error) {
	err := s.encoder.Encode(msg)
	// TODO: handle recoverable error...
	if err != nil {
		s.conn.Close()
		s.conn = nil
		return nil, err
	}

	if !msg.RequireReply() {
		return nil, nil
	}

	reply := NewEmptyMessage()
	err = s.decoder.Decode(reply)
	if err != nil {
		s.conn.Close()
		s.conn = nil
		return nil, err
	}
	return reply, nil
}

func (s *Sender) GoSend() {

}
