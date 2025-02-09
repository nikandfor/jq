package jq

import "nikand.dev/go/cbor"

type (
	Decoder struct {
		CBOR cbor.Decoder
	}
)

func MakeDecoder() Decoder { return Decoder{CBOR: cbor.MakeDecoder()} }

func (d Decoder) TagOnly(b []byte, st int) (tag Tag) {
	return Tag(b[st]) & cbor.TagMask
}

func (d Decoder) Tag(b []byte, st int) (tag Tag, sub int64, l, s, i int) {
	tag = Tag(b[st]) & cbor.TagMask

	if arrOrMap(tag) {
		tag, l, s, i = d.TagArrayMap(b, st)
	} else {
		tag, sub, i = d.CBOR.Tag(b, st)
	}

	return
}

func (d Decoder) UnderAllLabelsTagOnly(b []byte, st int) (tag Tag) {
	i := d.SkipAllLabels(b, st)

	return d.TagOnly(b, i)
}

func (d Decoder) UnderAllLabelsTag(b []byte, st int) (tag Tag, sub int64, l, s, i int) {
	i = d.SkipAllLabels(b, st)

	return d.Tag(b, i)
}

func (d Decoder) TagArrayMap(b []byte, st int) (tag Tag, l, s, i int) {
	tag = Tag(b[st]) & cbor.TagMask

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
	tag := Tag(b[st]) & cbor.TagMask

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
	tag := Tag(b[st]) & cbor.TagMask

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

func (d Decoder) SkipLabel(b []byte, st int) int {
	tag := d.TagOnly(b, st)
	if tag != cbor.Label {
		return st
	}

	_, i := d.Label(b, st)
	return i
}

func (d Decoder) SkipAllLabels(b []byte, st int) int {
	i := st

	for {
		tag := d.TagOnly(b, i)
		if tag != cbor.Label {
			return i
		}

		_, i = d.Label(b, i)
	}
}

func (d Decoder) Skip(b []byte, st int) int {
	i := d.SkipAllLabels(b, st)
	tag := d.TagOnly(b, i)

	if arrOrMap(tag) {
		_, l, s, i := d.TagArrayMap(b, i)
		return i + l*s
	}

	return d.CBOR.Skip(b, i)
}

func (d Decoder) Raw(b []byte, st int) ([]byte, int) {
	i := d.Skip(b, st)

	return b[st:i], i
}

func (d Decoder) Label(b []byte, st int) (int, int)       { return d.CBOR.Label(b, st) }
func (d Decoder) Signed(b []byte, st int) (int64, int)    { return d.CBOR.Signed(b, st) }
func (d Decoder) Unsigned(b []byte, st int) (uint64, int) { return d.CBOR.Unsigned(b, st) }
func (d Decoder) Bytes(b []byte, st int) ([]byte, int)    { return d.CBOR.Bytes(b, st) }
func (d Decoder) Float(b []byte, st int) (float64, int)   { return d.CBOR.Float(b, st) }
func (d Decoder) Float32(b []byte, st int) (float32, int) { return d.CBOR.Float32(b, st) }

func arrOrMap(tag Tag) bool {
	return tag&0b1100_0000 == 0b1000_0000
}
