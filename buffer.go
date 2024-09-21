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

func (b BufferReader) Tag(off Off) byte {
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

func (b BufferReader) Raw(off Off) []byte {
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

func (b BufferReader) Simple(off Off) int {
	switch off {
	case False, True, Null, None:
		q := []int{
			-None:  cbor.None,
			-Null:  cbor.Null,
			-True:  cbor.True,
			-False: cbor.False,
		}

		return q[-off]
	}

	_, sub, _ := b.Decoder.CBOR.Tag(b.Buf(off))
	return int(sub)
}

func (b BufferReader) Float(off Off) float64 {
	v, _ := b.Decoder.Float(b.Buf(off))
	return v
}

func (b BufferReader) Float32(off Off) float32 {
	v, _ := b.Decoder.Float32(b.Buf(off))
	return v
}

func (b BufferReader) IsSimple(off Off, specials ...Off) (ok bool) {
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
		x := []Off{
			0: -Zero,
			1: -One,
		}

		return int(sub) < len(x) && bits&(1<<x[sub]) != 0
	}

	if tag == cbor.Simple {
		x := []Off{
			cbor.None:  -None,
			cbor.Null:  -Null,
			cbor.True:  -True,
			cbor.False: -False,
		}

		return int(sub) < len(x) && bits&(1<<x[sub]) != 0
	}

	return false
}

func (b BufferReader) Bytes(off Off) []byte {
	s, _ := b.Decoder.Bytes(b.Buf(off))
	return s
}

func (b BufferReader) ArrayMapLen(off Off) int {
	_, l, _, _ := b.Decoder.TagArrayMap(b.Buf(off))
	return l
}

func (b BufferReader) ArrayMapIndex(off Off, index int) (k, v Off) {
	buf, base, eoff := b.BufBase(off)
	return b.Decoder.ArrayMapIndex(buf, base, eoff, index)
}

func (b BufferReader) ArrayMap(off Off, arr []Off) []Off {
	buf, base, eoff := b.BufBase(off)
	arr, _ = b.Decoder.ArrayMap(buf, base, eoff, arr)
	return arr
}

func (b BufferReader) Signed(off Off) int64 {
	switch off {
	case Zero:
		return 0
	case One:
		return 1
	}

	v, _ := b.Decoder.Signed(b.Buf(off))
	return v
}

func (b BufferReader) Unsigned(off Off) uint64 {
	switch off {
	case Zero:
		return 0
	case One:
		return 1
	}

	v, _ := b.Decoder.Unsigned(b.Buf(off))
	return v
}

func (b BufferWriter) Off() Off {
	return Off(len(b.R) + len(b.W))
}

func (b BufferWriter) Reset(off Off) {
	b.W = b.W[:int(off)-len(b.R)]
}

func (b BufferWriter) ResetIfErr(off Off, errp *error) {
	if *errp == nil {
		return
	}

	b.Reset(off)
}

func (b BufferWriter) Raw(raw []byte) Off {
	off := b.Off()
	b.W = append(b.W, raw...)

	return off
}

func (b BufferWriter) Array(arr []Off) Off {
	off := b.Off()
	b.W = b.Encoder.AppendArray(b.W, off, arr)
	return off
}

func (b BufferWriter) Map(arr []Off) Off {
	off := b.Off()
	b.W = b.Encoder.AppendMap(b.W, off, arr)
	return off
}

func (b BufferWriter) ArrayMap(tag byte, arr []Off) Off {
	off := b.Off()
	b.W = b.Encoder.AppendArrayMap(b.W, tag, off, arr)
	return off
}

func (b BufferWriter) String(v string) Off {
	off := b.Off()
	b.W = b.Encoder.AppendString(b.W, v)
	return off
}

func (b BufferWriter) Bytes(v []byte) Off {
	off := b.Off()
	b.W = b.Encoder.AppendBytes(b.W, v)
	return off
}

func (b BufferWriter) TagString(tag byte, v string) Off {
	off := b.Off()
	b.W = b.Encoder.AppendTagString(b.W, tag, v)
	return off
}

func (b BufferWriter) TagBytes(tag byte, v []byte) Off {
	off := b.Off()
	b.W = b.Encoder.AppendTagBytes(b.W, tag, v)
	return off
}

func (b BufferWriter) Int(v int) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.W = b.Encoder.AppendInt(b.W, v)
	return off
}

func (b BufferWriter) Int64(v int64) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.W = b.Encoder.AppendInt64(b.W, v)
	return off
}

func (b BufferWriter) Uint(v uint) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.W = b.Encoder.AppendUint(b.W, v)
	return off
}

func (b BufferWriter) Uint64(v uint64) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.W = b.Encoder.AppendUint64(b.W, v)
	return off
}

func (b BufferWriter) Float(v float64) Off {
	off := b.Off()
	b.W = b.Encoder.AppendFloat(b.W, v)
	return off
}

func (b BufferWriter) Float32(v float32) Off {
	off := b.Off()
	b.W = b.Encoder.AppendFloat32(b.W, v)
	return off
}

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
