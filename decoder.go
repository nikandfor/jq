package jq

import "nikand.dev/go/cbor"

type (
	Decoder struct {
		CBOR cbor.Decoder
	}
)

func (d Decoder) TagOnly(b []byte, off int) (tag byte) {
	return b[off] & cbor.TagMask
}

func (d Decoder) Tag(b []byte, off int) (tag byte, i int) {
	tag = b[off] & cbor.TagMask

	if tag&0b1100_0000 == 0b1000_0000 {
		tag, _, _, i = d.TagArrayMap(b, off)
	} else {
		tag, _, i = d.CBOR.Tag(b, off)
	}

	return tag, i
}

func (d Decoder) TagArrayMap(b []byte, off int) (tag byte, l, s, i int) {
	tag = b[off] & cbor.TagMask

	if b[off]&arrEmbedMask == 0 {
		l = int(b[off] & 0xf)
		s = 1
		i = int(off) + 1

		return tag, l, s, i
	}

	llss := b[off] & 0xf

	ls := 1 << (llss >> 2)
	s = 1 << (llss & 0b11)

	i = off + 1

	l = d.IntX(b, i, ls)
	i += ls

	return tag, l, s, i
}

func (d Decoder) ArrayMapIndex(b []byte, base, off, index int) (k, v int) {
	tag := b[off] & cbor.TagMask

	if b[off]&arrEmbedMask == 0 {
		l := int(b[off] & 0xf)
		if index < 0 {
			index = l + index
		}
		if index >= l || index < 0 {
			return None, None
		}

		i := int(off) + 1

		if tag == cbor.Map {
			return base + off - int(b[i+2*index]), base + off - int(b[i+2*index+1])
		}

		return None, base + off - int(b[i+index])
	}

	llss := b[off] & 0xf

	ls := 1 << (llss >> 2)
	s := 1 << (llss & 0b11)

	i := off + 1

	l := d.IntX(b, i, ls)
	i += ls

	if index < 0 {
		index = l + index
	}
	if index >= l || index < 0 {
		return None, None
	}

	if tag == cbor.Map {
		return base + off - d.IntX(b, i+2*s*index, s), base + off - d.IntX(b, 2*s*index+s, s)
	}

	return None, base + off - d.IntX(b, i+index, s)
}

func (d Decoder) ArrayMap(b []byte, base, off int, arr []int) ([]int, int) {
	tag := b[off] & cbor.TagMask

	if b[off]&arrEmbedMask == 0 {
		l := int(b[off] & 0xf)
		if tag == cbor.Map {
			l *= 2
		}

		i := off + 1

		for j := range l {
			arr = append(arr, base+off-int(b[i+j]))
		}

		return arr, i + l
	}

	llss := b[off] & 0xf

	ls := 1 << (llss >> 2)
	s := 1 << (llss & 0b11)

	i := off + 1

	l := d.IntX(b, i, ls)
	i += ls

	if tag == cbor.Map {
		l *= 2
	}

	for j := range l {
		arr = append(arr, base+off-d.IntX(b, i+s*j, s))
	}

	return arr, i + l*s
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
