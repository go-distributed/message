package message

import (
	"code.google.com/p/gogoprotobuf/proto"
)

type PbMessage struct {
	// msgType 0-127 do not require reply
	// msgType 128-255 require reply
	msgType uint8
	pb      proto.Message
	reply   chan *PbMessage
}

func NewPbMessage(msgType uint8, pb proto.Message) *PbMessage {
	m := &PbMessage{
		msgType: msgType,
		pb:      pb,
	}

	return m
}

func NewEmptyPbMessage() *PbMessage {
	return NewPbMessage(0, nil)
}

func (m *PbMessage) Type() uint8 { return m.msgType }

func (m *PbMessage) Proto() proto.Message { return m.pb }

func (m *PbMessage) AttachReplyChan() bool {
	if m.msgType > MsgRequireReply {
		m.reply = make(chan *PbMessage, 1)
		return true
	}
	return false
}

func (m *PbMessage) RequireReply() bool {
	return m.msgType > MsgRequireReply
}
