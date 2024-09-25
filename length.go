package jq

import (
	"unicode/utf8"

	"nikand.dev/go/cbor"
)

type (
	Length struct {
		CountRunes bool
	}
)

func NewLength() Length { return Length{} }

func (f Length) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	te := NewTypeError(b.Reader().TagRaw(off))

	if off < 0 {
		switch off {
		case Zero, One:
			return off, false, nil
		case Null, EmptyString, EmptyArray:
			return Zero, false, nil
		default:
			return off, false, te
		}
	}

	tag, sub, l, _, _ := b.Decoder.Tag(b.Buf(off))

	switch tag {
	case cbor.Int:
		res = off
	case cbor.Neg:
		v := b.Reader().Unsigned(off)
		res = b.Writer().Uint64(v)
	case cbor.Bytes:
		res = b.Writer().Int(int(sub))
	case cbor.String:
		if f.CountRunes {
			s := b.Reader().Bytes(off)
			res = b.Writer().Int(utf8.RuneCount(s))
		} else {
			res = b.Writer().Int(int(sub))
		}
	case cbor.Array, cbor.Map:
		res = b.Writer().Int(l)
	case cbor.Simple:
		switch sub {
		case cbor.Null:
			res = Zero
		case cbor.Float8, cbor.Float16, cbor.Float32, cbor.Float64:
			f := b.Reader().Float(off)
			if f >= 0 {
				res = off
			} else {
				res = b.Writer().Float(-f)
			}
		default:
			return off, false, te
		}
	default:
		return off, false, te
	}

	return res, false, nil
}

func (f Length) String() string {
	if f.CountRunes {
		return "length_utf8"
	}

	return "length"
}
