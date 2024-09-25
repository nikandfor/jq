package jq

import (
	"bytes"

	"nikand.dev/go/cbor"
)

type (
	Buffer struct {
		B []byte

		Decoder Decoder
		Encoder Encoder
	}

	BufferReader struct {
		*Buffer
	}
	BufferWriter struct {
		*Buffer
	}
)

var (
	shortToCBOR = []Tag{
		-None:  cbor.Simple | cbor.None,
		-Null:  cbor.Simple | cbor.Null,
		-True:  cbor.Simple | cbor.True,
		-False: cbor.Simple | cbor.False,
		-Zero:  cbor.Int | 0,
		-One:   cbor.Int | 1,

		-EmptyString: cbor.String | 0,
		-EmptyArray:  cbor.Array | 0,
	}

	cborToShort = []Off{
		cbor.Simple | cbor.Null:  Null,
		cbor.Simple | cbor.True:  True,
		cbor.Simple | cbor.False: False,
		cbor.Int | 0:             Zero,
		cbor.Int | 1:             One,

		cbor.String | 0: EmptyString,
		cbor.Array | 0:  EmptyArray,
	}
)

func NewBuffer() *Buffer {
	b := MakeBuffer()
	return &b
}

func MakeBuffer() Buffer {
	return Buffer{
		Encoder: MakeEncoder(),
		Decoder: MakeDecoder(),
	}
}

func (b *Buffer) Reset() {
	b.B = b.B[:0]
}

func (b *Buffer) Reader() BufferReader { return BufferReader{b} }
func (b *Buffer) Writer() BufferWriter { return BufferWriter{b} }

func (b *Buffer) Equal(loff, roff Off) (res bool) {
	br := b.Reader()

	//	log.Printf("equal %x %x\n%s", loff, roff, DumpBuffer(b))
	//	defer func() { log.Printf("equal %x %x  =>  %v", loff, roff, res) }()

	if loff == roff {
		return true
	}

	tag := br.Tag(loff)
	rtag := br.Tag(roff)

	if tag != rtag {
		return false
	}

	switch tag {
	case cbor.Int, cbor.Neg, cbor.Bytes, cbor.String, cbor.Simple, cbor.Labeled:
		lraw := br.Raw(loff)
		rraw := br.Raw(roff)

		return bytes.Equal(lraw, rraw)
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	larr := br.ArrayMap(loff, nil)
	rarr := br.ArrayMap(roff, nil)

	if len(larr) != len(rarr) {
		return false
	}

	for i := range larr {
		if !b.Equal(larr[i], rarr[i]) {
			return false
		}
	}

	return true
}

func (b *Buffer) Len() int {
	if b == nil {
		return 0
	}

	return len(b.B)
}

func (b *Buffer) Buf(off Off) ([]byte, int) {
	return b.B, int(off)
}

func (b *Buffer) Unwrap() []byte {
	return b.B
}

func (b *Buffer) Dump() string {
	return NewDumper(nil).Dump(b)
}
