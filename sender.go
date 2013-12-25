package message

type Sender struct {
}

func (s *Sender) Send() *Message {
	return NewEmptyMessage()
}

func (s *Sender) GoSend() {

}
