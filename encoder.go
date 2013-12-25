package message

import (
	"bufio"
	"encoding/binary"
	"io"
)

type MsgEncoder struct {
	bw *bufio.Writer
}

func NewMsgEncoder(w io.Writer) *MsgEncoder {
	return &MsgEncoder{
		bw: bufio.NewWriter(w),
	}
}

func (me *MsgEncoder) Encode(m *Message) error {
	err := me.bw.WriteByte(byte(m.msgType))
	if err != nil {
		panic(err)
	}

	size := len(m.bytes)
	err = binary.Write(me.bw, binary.LittleEndian, uint32(size))

	if err != nil {
		panic(err)
	}

	_, err = me.bw.Write(m.bytes)

	return me.bw.Flush()
}
