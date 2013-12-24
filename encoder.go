package message

import (
	"encoding/binary"
	"io"

	"code.google.com/p/gogoprotobuf/proto"
)

type PbEncoder struct {
	w   io.Writer
	buf *proto.Buffer
}

func NewPbEecoder(w io.Writer) *PbEncoder {
	return &PbEncoder{
		w:   w,
		buf: proto.NewBuffer(nil),
	}
}

func (e *PbEncoder) Encode(pb proto.Message) (int, error) {
	err := e.buf.Marshal(pb)

	// TODO: deal with err
	if err != nil {
		panic(err)
	}

	size := len(e.buf.Bytes())
	err = binary.Write(e.w, binary.LittleEndian, uint32(size))
	// TODO: deal with err
	if err != nil {
		panic(err)
	}

	return e.w.Write(e.buf.Bytes())
}
