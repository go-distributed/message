package message

const (
	MsgRequireReply = 127
)

type Message struct {
	// msgType 0-127 do not require reply
	// msgType 128-255 require reply
	msgType uint8
	bytes   []byte
	reply   chan *Message
}

func NewMessage(msgType uint8, bytes []byte) *Message {
	m := &Message{
		msgType: msgType,
		bytes:   bytes,
	}

	return m
}

func NewEmptyMessage() *Message {
	return NewMessage(0, nil)
}

func (m *Message) Type() uint8 { return m.msgType }

func (m *Message) Bytes() []byte { return m.bytes }

func (m *Message) AttachReplyChan() bool {
	if m.msgType > MsgRequireReply {
		m.reply = make(chan *Message, 1)
		return true
	}
	return false
}
