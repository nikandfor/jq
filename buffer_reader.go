package jq

import "nikand.dev/go/cbor"

func (b BufferReader) Tag(off Off) byte {
	if off < 0 {
		return shortToCBOR[-off] & cbor.TagMask
	}

	tag, _, _, _, _ := b.Decoder.Tag(b.Buf(off))
	return tag
}

func (b BufferReader) TagRaw(off Off) byte {
	if off < 0 {
		return shortToCBOR[-off]
	}

	buf, st := b.Buf(off)

	return buf[st]
}

func (b BufferReader) Raw(off Off) []byte {
	if off < 0 {
		return []byte{shortToCBOR[-off]}
	}

	raw, _ := b.Decoder.Raw(b.Buf(off))
	return raw
}

func (b BufferReader) Simple(off Off) int {
	if off < 0 {
		return int(shortToCBOR[-off]) & cbor.SubMask
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

func (b BufferReader) String(off Off) string {
	s, _ := b.Decoder.Bytes(b.Buf(off))
	return string(s)
}

func (b BufferReader) ArrayMapLen(off Off) int {
	if off == EmptyArray {
		return 0
	}

	_, l, _, _ := b.Decoder.TagArrayMap(b.Buf(off))
	return l
}

func (b BufferReader) ArrayMapIndex(off Off, index int) (k, v Off) {
	if off == EmptyArray {
		return None, Null
	}

	buf, st := b.Buf(off)
	return b.Decoder.ArrayMapIndex(buf, st, index)
}

func (b BufferReader) ArrayMap(off Off, arr []Off) []Off {
	if off == EmptyArray {
		return arr
	}

	buf, st := b.Buf(off)
	arr, _ = b.Decoder.ArrayMap(buf, st, arr)
	return arr
}

func (b BufferReader) Int(off Off) int {
	return int(b.Signed(off))
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
