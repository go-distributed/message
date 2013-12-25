package message

type Message struct {
	msgType uint8
	bytes   []byte
}

func NewMessage(msgType uint8, bytes []byte) *Message {
	return &Message{
		msgType: msgType,
		bytes:   bytes,
	}
}

func NewEmptyMessage() *Message {
	return NewMessage(0, nil)
}

func (m *Message) Type() uint8 { return m.msgType }

func (m *Message) Bytes() []byte { return m.bytes }
