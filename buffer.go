package jq

import (
	"bytes"

	"nikand.dev/go/cbor"
)

type (
	Buffer struct {
		R, W []byte

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

var shortToCBOR = []byte{
	-None:  cbor.Simple | cbor.None,
	-Null:  cbor.Simple | cbor.Null,
	-True:  cbor.Simple | cbor.True,
	-False: cbor.Simple | cbor.False,
	-Zero:  cbor.Int | 0,
	-One:   cbor.Int | 1,

	-EmptyString: cbor.String | 0,
	-EmptyArray:  cbor.Array | 0,
}

func NewBuffer(r []byte) *Buffer {
	return &Buffer{R: r}
}

func MakeBuffer(r []byte) Buffer {
	return Buffer{R: r}
}

func (b *Buffer) Reset(r []byte) {
	b.R = r
	b.W = b.W[:0]
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

func (b *Buffer) Buf(off Off) ([]byte, int) {
	if int(off) < len(b.R) {
		return b.R, int(off)
	}

	return b.W, int(off) - len(b.R)
}

func (b *Buffer) BufBase(off Off) ([]byte, int, int) {
	if int(off) < len(b.R) {
		return b.R, 0, int(off)
	}

	return b.W, len(b.R), int(off) - len(b.R)
}

func (b *Buffer) Unwrap() (r0, r1 []byte) {
	return b.R, b.W
}

func (b *Buffer) Shift() {
	b.R, b.W = b.W, nil
}

func (b *Buffer) Unshift() {
	b.R, b.W = nil, b.R
}
