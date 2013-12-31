package message

import (
	"bufio"
	"encoding/binary"
	"io"

	"code.google.com/p/gogoprotobuf/proto"
)

type MsgEncoder struct {
	bw   *bufio.Writer
	pbuf *proto.Buffer
}

func NewMsgEncoder(w io.Writer) *MsgEncoder {
	return &MsgEncoder{
		bw:   bufio.NewWriter(w),
		pbuf: proto.NewBuffer(nil),
	}
}

func (me *MsgEncoder) EncodePb(m *PbMessage) error {
	err := me.bw.WriteByte(byte(m.msgType))
	if err != nil {
		return err
	}

	var bytes []byte
	if m.pb != nil {
		me.pbuf.Reset()
		err = me.pbuf.Marshal(m.pb)
		if err != nil {
			return err
		}
		bytes = me.pbuf.Bytes()
	}
	size := len(bytes)

	err = binary.Write(me.bw, binary.LittleEndian, uint32(size))

	if err != nil {
		return err
	}

	_, err = me.bw.Write(bytes)

	return me.bw.Flush()
}

func (me *MsgEncoder) Encode(m *Message) error {
	err := me.bw.WriteByte(byte(m.msgType))
	if err != nil {
		return err
	}

	size := len(m.bytes)
	err = binary.Write(me.bw, binary.LittleEndian, uint32(size))

	if err != nil {
		return err
	}

	_, err = me.bw.Write(m.bytes)

	return me.bw.Flush()
}
