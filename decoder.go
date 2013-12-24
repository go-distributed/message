package message

import (
	"encoding/binary"
	"io"

	"code.google.com/p/gogoprotobuf/proto"
)

type PbDecoder struct {
	r   io.Reader
	buf *proto.Buffer
}

func NewPbDecoder(r io.Reader) *PbDecoder {
	return &PbDecoder{
		r:   r,
		buf: proto.NewBuffer(nil),
	}
}

func (d *PbDecoder) Decode(pb proto.Message) error {
	var size uint32
	err := binary.Read(d.r, binary.LittleEndian, &size)
	// TODO: deal with error...
	if err != nil {
		panic(err)
	}

	buf := make([]byte, size)
	_, err = io.ReadFull(d.r, buf)
	// TODO: deal with error...
	if err != nil {
		panic(err)
	}

	d.buf.SetBuf(buf)

	return d.buf.Unmarshal(pb)
}
