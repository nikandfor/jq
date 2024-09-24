package jq

import (
	"nikand.dev/go/cbor"
)

type (
	Decoder struct {
		CBOR cbor.Decoder
	}
)

func (d Decoder) TagOnly(b []byte, off int) (tag byte) {
	return b[off] & cbor.TagMask
}

func (d Decoder) Tag(b []byte, off int) (tag byte, sub int64, l, s, i int) {
	tag = b[off] & cbor.TagMask

	if tag&0b1100_0000 == 0b1000_0000 {
		tag, l, s, i = d.TagArrayMap(b, off)
	} else {
		tag, sub, i = d.CBOR.Tag(b, off)
	}

	return
}

func (d Decoder) TagArrayMap(b []byte, st int) (tag byte, l, s, i int) {
	tag = b[st] & cbor.TagMask

	if b[st]&arrEmbedMask == 0 {
		l = int(b[st] & 0xf)
		s = 1
		i = st + 1

		return tag, l, s, i
	}

	llss := b[st] & 0xf

	ls := 1 << (llss >> 2)
	s = 1 << (llss & 0b11)

	i = st + 1

	l = d.IntX(b, i, ls)
	i += ls

	return tag, l, s, i
}

func (d Decoder) ArrayMapIndex(b []byte, st, index int) (k, v Off) {
	tag := b[st] & cbor.TagMask

	if b[st]&arrEmbedMask == 0 {
		l := int(b[st] & 0xf)
		if index < 0 {
			index = l + index
		}
		if index >= l || index < 0 {
			return None, None
		}

		i := st + 1

		val := func(j int) Off {
			if b[i+j] >= 0x100-offReserve {
				return Off(b[i+j]) - 0x100
			}

			return Off(st) - Off(b[i+j])
		}

		if tag == cbor.Map {
			return val(2 * index), val(2*index + 1)
		}

		return None, val(index)
	}

	llss := b[st] & 0xf

	ls := 1 << (llss >> 2)
	ss := 1 << (llss & 0b11)

	i := st + 1

	l := d.IntX(b, i, ls)
	i += ls

	if index < 0 {
		index = l + index
	}
	if index >= l || index < 0 {
		return None, None
	}

	size := 1 << (8 * ss)

	val := func(j int) Off {
		v := d.IntX(b, i+j*ss, ss)
		if v >= size-offReserve {
			return Off(v - size)
		}

		return Off(st - v)
	}

	if tag == cbor.Map {
		return val(2 * index), val(2*index + 1)
	}

	return None, val(index)
}

func (d Decoder) ArrayMap(b []byte, st int, arr []Off) ([]Off, int) {
	tag := b[st] & cbor.TagMask

	if b[st]&arrEmbedMask == 0 {
		l := int(b[st] & 0xf)
		if tag == cbor.Map {
			l *= 2
		}

		i := st + 1

		val := func(j int) Off {
			if b[i+j] >= 0x100-offReserve {
				return Off(b[i+j]) - 0x100
			}

			return Off(st) - Off(b[i+j])
		}

		for j := range l {
			arr = append(arr, val(j))
		}

		//	log.Printf("decode array/map  %x %x+%x -> % x", tag, base, off, arr)

		return arr, i + l
	}

	llss := b[st] & 0xf

	ls := 1 << (llss >> 2)
	ss := 1 << (llss & 0b11)

	i := st + 1

	l := d.IntX(b, i, ls)
	i += ls

	if tag == cbor.Map {
		l *= 2
	}

	size := 1 << (8 * ss)

	val := func(j int) Off {
		v := d.IntX(b, i+j*ss, ss)
		if v >= size-offReserve {
			return Off(v - size)
		}

		return Off(st) - Off(v)
	}

	for j := range l {
		arr = append(arr, val(j))
	}

	return arr, i + l*ss
}

func (d Decoder) IntX(b []byte, i, x int) int {
	switch x {
	case 1:
		return int(b[i])
	case 2:
		return int(b[i])<<8 | int(b[i+1])
	case 4:
		return int(b[i])<<24 | int(b[i+1])<<16 | int(b[i+2])<<8 | int(b[i+3])
	default:
		panic("overflow")
	}
}

func (d Decoder) Skip(b []byte, st int) int {
	tag := b[st] & cbor.TagMask

	if tag&0b1100_0000 == 0b1000_0000 {
		_, l, s, i := d.TagArrayMap(b, st)
		return i + l*s
	}

	return d.CBOR.Skip(b, st)
}

func (d Decoder) Raw(b []byte, st int) ([]byte, int) {
	tag := b[st] & cbor.TagMask

	if tag&0b1100_0000 == 0b1000_0000 {
		_, l, s, i := d.TagArrayMap(b, st)
		i += l * s
		return b[st:i], i
	}

	return d.CBOR.Raw(b, st)
}
func (d Decoder) Signed(b []byte, st int) (int64, int)    { return d.CBOR.Signed(b, st) }
func (d Decoder) Unsigned(b []byte, st int) (uint64, int) { return d.CBOR.Unsigned(b, st) }
func (d Decoder) Bytes(b []byte, st int) ([]byte, int)    { return d.CBOR.Bytes(b, st) }
func (d Decoder) Float(b []byte, st int) (float64, int)   { return d.CBOR.Float(b, st) }
func (d Decoder) Float32(b []byte, st int) (float32, int) { return d.CBOR.Float32(b, st) }
