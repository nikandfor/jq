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
	case False, True, Null:
		return cbor.Simple
	case Zero, One:
		return cbor.Int
	}

	tag, _, _, _, _ := b.Decoder.Tag(b.Buf(off))
	return tag
}

func (b BufferReader) Raw(off int) []byte {
	switch off {
	case False, True, Null, None:
		q := []byte{
			-None:  cbor.None,
			-Null:  cbor.Null,
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

func (b BufferReader) Simple(off int) int {
	_, sub, _ := b.Decoder.CBOR.Tag(b.Buf(off))
	return int(sub)
}

func (b BufferReader) Float(off int) float64 {
	v, _ := b.Decoder.Float(b.Buf(off))
	return v
}

func (b BufferReader) Float32(off int) float32 {
	v, _ := b.Decoder.Float32(b.Buf(off))
	return v
}

func (b BufferReader) IsSimple(off int, specials ...int) (ok bool) {
	if off < 0 {
		for _, v := range specials {
			ok = ok || v == off
		}

		return ok
	}

	bits := 0

	for _, v := range specials {
		bits |= 1 << -v
	}

	tag, sub, _, _, _ := b.Decoder.Tag(b.Buf(off))

	if tag == cbor.Int {
		x := []int{
			0: -Zero,
			1: -One,
		}

		return int(sub) < len(x) && bits&(1<<x[sub]) != 0
	}

	if tag == cbor.Simple {
		x := []int{
			cbor.None:  -None,
			cbor.Null:  -Null,
			cbor.True:  -True,
			cbor.False: -False,
		}

		return int(sub) < len(x) && bits&(1<<x[sub]) != 0
	}

	return false
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

func (b BufferReader) Signed(off int) int64 {
	switch off {
	case Zero:
		return 0
	case One:
		return 1
	}

	v, _ := b.Decoder.Signed(b.Buf(off))
	return v
}

func (b BufferReader) Unsigned(off int) uint64 {
	switch off {
	case Zero:
		return 0
	case One:
		return 1
	}

	v, _ := b.Decoder.Unsigned(b.Buf(off))
	return v
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

func (b BufferWriter) ArrayMap(tag byte, arr []int) int {
	off := b.Len()
	b.W = b.Encoder.AppendArrayMap(b.W, tag, off, arr)
	return off
}

func (b BufferWriter) String(v string) int {
	off := b.Len()
	b.W = b.Encoder.AppendString(b.W, v)
	return off
}

func (b BufferWriter) Bytes(v []byte) int {
	off := b.Len()
	b.W = b.Encoder.AppendBytes(b.W, v)
	return off
}

func (b BufferWriter) TagString(tag byte, v string) int {
	off := b.Len()
	b.W = b.Encoder.AppendTagString(b.W, tag, v)
	return off
}

func (b BufferWriter) TagBytes(tag byte, v []byte) int {
	off := b.Len()
	b.W = b.Encoder.AppendTagBytes(b.W, tag, v)
	return off
}

func (b BufferWriter) Int(v int) int {
	off := b.Len()
	b.W = b.Encoder.AppendInt(b.W, v)
	return off
}

func (b BufferWriter) Int64(v int64) int {
	off := b.Len()
	b.W = b.Encoder.AppendInt64(b.W, v)
	return off
}

func (b BufferWriter) Uint(v uint) int {
	off := b.Len()
	b.W = b.Encoder.AppendUint(b.W, v)
	return off
}

func (b BufferWriter) Uint64(v uint64) int {
	off := b.Len()
	b.W = b.Encoder.AppendUint64(b.W, v)
	return off
}

func (b BufferWriter) Float(v float64) int {
	off := b.Len()
	b.W = b.Encoder.AppendFloat(b.W, v)
	return off
}

func (b BufferWriter) Float32(v float32) int {
	off := b.Len()
	b.W = b.Encoder.AppendFloat32(b.W, v)
	return off
}

func (b *Buffer) Equal(loff int, roff int) (res bool) {
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

func (b *Buffer) Buf(off int) ([]byte, int) {
	if off < len(b.R) {
		return b.R, off
	}

	return b.W, off - len(b.R)
}

func (b *Buffer) BufBase(off int) ([]byte, int, int) {
	if off < len(b.R) {
		return b.R, 0, off
	}

	return b.W, len(b.R), off - len(b.R)
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
