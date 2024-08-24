package jq

import "nikand.dev/go/cbor"

type (
	Encoder struct {
		CBOR cbor.Encoder
	}
)

const (
	// 0bxxx0_yyyy // xxx - Array or Map, yyyy (count) of 1-byte offsets (or pairs)
	// 0bxxx1_llss // ll - length size, ss - offset size
	//             // 1<<ll elements of size 1<<ss bytes are followed. 1<<(ll+1) for Map

	arrEmbedMask = 0b0001_0000

	offReserve = 8
)

func (e Encoder) AppendArray(b []byte, off int, items []int) []byte {
	return e.AppendArrayMap(b, cbor.Array, off, items)
}

func (e Encoder) AppendMap(b []byte, off int, items []int) []byte {
	return e.AppendArrayMap(b, cbor.Map, off, items)
}

func (e Encoder) AppendArrayMap(b []byte, tag byte, off int, items []int) []byte {
	tagLen := len(items)
	if tag == cbor.Map {
		tagLen /= 2
	}

	min := 0

	for _, item := range items {
		if item < min {
			min = item
		}
	}

	d := off - min

	if d < 0x100 && tagLen < 16 {
		b = append(b, tag|byte(tagLen))

		for _, it := range items {
			b = append(b, byte(off-it))
		}

		return b
	}

	var ll, ss byte = 0, 0

	for q := 0x100; tagLen >= q; {
		ll++
		q = q * q
	}

	for q := 0x100; d >= q; {
		ss++
		q = q * q
	}

	f := tag | ll<<2 | ss

	b = append(b, f)
	b = e.AppendIntX(b, 1<<ll, tagLen)

	for _, item := range items {
		b = e.AppendIntX(b, 1<<ss, off-item)
	}

	return b
}

func (e Encoder) AppendIntX(b []byte, x, v int) []byte {
	switch x {
	case 1:
		return append(b, byte(v))
	case 2:
		return append(b, byte(v>>8), byte(v))
	case 4:
		return append(b, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	default:
		panic("overflow")
	}
}

func (e Encoder) AppendRaw(b, raw []byte) []byte           { return append(b, raw...) }
func (e Encoder) AppendInt(b []byte, v int) []byte         { return e.CBOR.AppendInt(b, v) }
func (e Encoder) AppendInt64(b []byte, v int64) []byte     { return e.CBOR.AppendInt64(b, v) }
func (e Encoder) AppendUint(b []byte, v uint) []byte       { return e.CBOR.AppendUint(b, v) }
func (e Encoder) AppendUint64(b []byte, v uint64) []byte   { return e.CBOR.AppendUint64(b, v) }
func (e Encoder) AppendFloat(b []byte, v float64) []byte   { return e.CBOR.AppendFloat(b, v) }
func (e Encoder) AppendFloat32(b []byte, v float32) []byte { return e.CBOR.AppendFloat32(b, v) }
func (e Encoder) AppendBool(b []byte, v bool) []byte       { return e.CBOR.AppendBool(b, v) }
func (e Encoder) AppendNull(b []byte) []byte               { return e.CBOR.AppendNull(b) }
func (e Encoder) AppendString(b []byte, v string) []byte   { return e.CBOR.AppendString(b, v) }
func (e Encoder) AppendBytes(b []byte, v []byte) []byte    { return e.CBOR.AppendBytes(b, v) }

func (e Encoder) AppendTagString(b []byte, tag byte, v string) []byte {
	return e.CBOR.AppendTagString(b, tag, v)
}

func (e Encoder) AppendTagBytes(b []byte, tag byte, v []byte) []byte {
	return e.CBOR.AppendTagBytes(b, tag, v)
}
