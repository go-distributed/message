package message

import (
	"net"
)

type PbSender struct {
	remoteAddr *net.TCPAddr
	conn       *net.TCPConn
	encoder    *MsgEncoder
	decoder    *MsgDecoder
}

func NewPbSender(raddrStr string) (*PbSender, error) {
	raddr, err := net.ResolveTCPAddr("tcp", raddrStr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}

	return &PbSender{
		remoteAddr: raddr,
		conn:       conn,
		encoder:    NewMsgEncoder(conn),
		decoder:    NewMsgDecoder(conn),
	}, nil
}

func (s *PbSender) Send(msg *PbMessage) (*PbMessage, error) {
	err := s.encoder.EncodePb(msg)
	// TODO: handle recoverable error...
	if err != nil {
		s.conn.Close()
		s.conn = nil
		return nil, err
	}

	if !msg.RequireReply() {
		return nil, nil
	}

	reply := NewEmptyPbMessage()
	err = s.decoder.DecodePb(reply)
	if err != nil {
		s.conn.Close()
		s.conn = nil
		return nil, err
	}
	return reply, nil
}

func (s *PbSender) GoSend() {

}
