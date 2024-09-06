package jq

import (
	"unicode/utf8"

	"nikand.dev/go/cbor"
)

type (
	Length struct {
		LenRunes bool
	}
)

func (f Length) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	if off < 0 {
		switch off {
		case Zero, One:
			return off, false, nil
		case Null:
			return Zero, false, nil
		default:
			return off, false, ErrType
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
		if f.LenRunes {
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
			return off, false, ErrType
		}
	default:
		return off, false, ErrType
	}

	return res, false, nil
}

func (f Length) String() string { return "length" }
