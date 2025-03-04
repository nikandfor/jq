package jq

import "nikand.dev/go/cbor"

type (
	Encoder struct {
		CBOR cbor.Emitter
	}
)

const (
	// 0bxxx0_yyyy // xxx - Array or Map, yyyy (count) of 1-byte offsets (or pairs)
	// 0bxxx1_llss // ll - length size, ss - offset size
	//             // 1<<ll elements of size 1<<ss bytes are followed. 1<<(ll+1) for Map

	arrEmbedMask = 0b0001_0000
)

func MakeEncoder() Encoder { return Encoder{CBOR: cbor.MakeEmitter()} }

func (e Encoder) AppendArray(b []byte, off Off, items []Off) []byte {
	return e.AppendArrayMap(b, cbor.Array, off, items)
}

func (e Encoder) AppendMap(b []byte, off Off, items []Off) []byte {
	return e.AppendArrayMap(b, cbor.Map, off, items)
}

func (e Encoder) AppendArrayMap(b []byte, tag Tag, off Off, items []Off) []byte {
	//	reset := len(b)

	tagLen := len(items)
	if tag == cbor.Map {
		tagLen /= 2
	}

	min := off

	for _, item := range items {
		if item < 0 {
			continue
		}
		if item < min {
			min = item
		}
	}

	d := off - min
	size := 0x100

	if d < 0x100-offReserve && tagLen < 16 {
		b = append(b, byte(tag)|byte(tagLen))

		for _, item := range items {
			if item < 0 {
				b = append(b, byte(size)+byte(item))
				continue
			}

			b = append(b, byte(off-item))
		}

		//	log.Printf("append array/map  %x %x  % x -> 0 0   % x", tag, off, items, b[reset:])

		return b
	}

	var ll, ss byte = 0, 0

	for size = 0x100; tagLen >= size; {
		ll++
		size *= size
	}

	for size = 0x100; int(d) >= size-offReserve; {
		ss++
		size *= size
	}

	t := byte(tag) | 0b0001_0000 | ll<<2 | ss

	b = append(b, t)
	b = e.AppendIntX(b, 1<<ll, tagLen)

	for _, item := range items {
		if item < 0 {
			b = e.AppendIntX(b, 1<<ss, size+int(item))
			continue
		}

		b = e.AppendIntX(b, 1<<ss, int(off)-int(item))
	}

	//	log.Printf("append array/map  %x %x  % x -> %d %d   % x", tag, off, items, ll, ss, b[reset:])

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

func (e Encoder) AppendLabel(b []byte, lab int) []byte { return e.CBOR.AppendLabel(b, lab) }

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

func (e Encoder) AppendTagString(b []byte, tag Tag, v string) []byte {
	return e.CBOR.AppendTagString(b, tag, v)
}

func (e Encoder) AppendTagBytes(b []byte, tag Tag, v []byte) []byte {
	return e.CBOR.AppendTagBytes(b, tag, v)
}

func (e Encoder) AppendTagUnsigned(b []byte, tag Tag, v uint64) []byte {
	return e.CBOR.AppendTagUnsigned(b, tag, v)
}
