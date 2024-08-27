package jq

import "nikand.dev/go/cbor"

type (
	Length struct{}
)

func (f Length) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if next {
		return None, false, nil
	}

	if off < 0 {
		return off, false, ErrType
	}

	tag, sub, l, _, _ := b.Decoder.Tag(b.Buf(off))

	switch tag {
	case cbor.Bytes, cbor.String:
		res = b.Writer().Int(int(sub))
	case cbor.Array, cbor.Map:
		res = b.Writer().Int(l)
	default:
		return off, false, ErrType
	}

	return res, false, nil
}

func (f Length) String() string { return "length" }
