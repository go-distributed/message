package message

import (
	"bufio"
	"encoding/binary"
	"io"
	"reflect"

	"code.google.com/p/gogoprotobuf/proto"
)

type MsgDecoder struct {
	br *bufio.Reader
}

func NewMsgDecoder(r io.Reader) *MsgDecoder {
	return &MsgDecoder{
		br: bufio.NewReader(r),
	}
}

func (md *MsgDecoder) DecodePb(m *PbMessage) error {
	msgType, err := md.br.ReadByte()

	if err != nil {
		return err
	}

	m.msgType = uint8(msgType)

	t, ok := registry[m.msgType]

	if !ok {
		panic("unknown type")
	}

	v := reflect.New(t)
	m.pb = v.Interface().(proto.Message)

	var size uint32
	err = binary.Read(md.br, binary.LittleEndian, &size)
	if err != nil {
		return err
	}

	bytes := make([]byte, size)
	_, err = io.ReadFull(md.br, bytes)

	return proto.Unmarshal(bytes, m.pb)
}

func (md *MsgDecoder) Decode(m *Message) error {
	msgType, err := md.br.ReadByte()

	if err != nil {
		return err
	}

	m.msgType = uint8(msgType)

	var size uint32
	err = binary.Read(md.br, binary.LittleEndian, &size)
	if err != nil {
		return err
	}

	m.bytes = make([]byte, size)
	_, err = io.ReadFull(md.br, m.bytes)
	return err
}
