package message

type Receiver struct {
	ch chan *Message
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

func (r *Receiver) start() {

}
