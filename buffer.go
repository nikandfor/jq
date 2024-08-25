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

func (b BufferReader) Tag(off int) byte {
	switch off {
	case None:
		return cbor.Simple
	case False, True, Nil:
		return cbor.Simple
	case Zero, One:
		return cbor.Int
	}

	tag, _, _, _, _ := b.Decoder.Tag(b.Buf(off))
	return tag
}

func (b BufferReader) Raw(off int) []byte {
	switch off {
	case False, True, Nil, None:
		q := []byte{
			-None:  cbor.None,
			-Nil:   cbor.Null,
			-True:  cbor.True,
			-False: cbor.False,
		}

		return []byte{cbor.Simple | q[-off]}
	case Zero, One:
		return []byte{cbor.Int | byte(Zero-off)}
	}

	raw, _ := b.Decoder.Raw(b.Buf(off))
	return raw
}

func (b BufferReader) Bytes(off int) []byte {
	s, _ := b.Decoder.Bytes(b.Buf(off))
	return s
}

func (b BufferReader) ArrayMapIndex(off, index int) (k, v int) {
	buf, base, eoff := b.BufBase(off)
	return b.Decoder.ArrayMapIndex(buf, base, eoff, index)
}

func (b BufferReader) ArrayMap(off int, arr []int) []int {
	buf, base, eoff := b.BufBase(off)
	arr, _ = b.Decoder.ArrayMap(buf, base, eoff, arr)
	return arr
}

func (b BufferWriter) Len() int {
	return len(b.R) + len(b.W)
}

func (b BufferWriter) Reset(off int) {
	b.W = b.W[:off-len(b.R)]
}

func (b BufferWriter) ResetIfErr(off int, errp *error) {
	if *errp == nil {
		return
	}

	b.Reset(off)
}

func (b BufferWriter) Raw(raw []byte) int {
	off := b.Len()
	b.W = append(b.W, raw...)

	return off
}

func (b BufferWriter) Array(arr []int) int {
	off := b.Len()
	b.W = b.Encoder.AppendArray(b.W, off, arr)
	return off
}

func (b BufferWriter) Map(arr []int) int {
	off := b.Len()
	b.W = b.Encoder.AppendMap(b.W, off, arr)
	return off
}

func (b *Buffer) Equal(loff int, roff int) (res bool) {
	br := b.Reader()

	//	log.Printf("equal %x %x", loff, roff)
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

func (b *Buffer) Buf(off int) ([]byte, int) {
	if off < len(b.R) {
		return b.R, off
	}

	return b.W, off - len(b.R)
}

func (b *Buffer) BufBase(off int) ([]byte, int, int) {
	if off <= len(b.R) {
		return b.R, 0, off
	}

	return b.W, len(b.R), off - len(b.R)
}

func (b *Buffer) Unwrap() (r0, r1 []byte) {
	return b.R, b.W
}
