package jqcbor

import (
	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Encoder struct {
		CBOR cbor.Encoder
		JQ   jq.Decoder

		arr []int
	}
)

func (e *Encoder) Encode(w, r []byte, off int) []byte {
	tag, sub, i := e.JQ.Decoder.Tag(r, off)

	switch tag {
	case cbor.Int, cbor.Neg, cbor.Bytes, cbor.String:
		if tag == cbor.Bytes || tag == cbor.String {
			i += int(sub)
		}

		return append(w, r[off:i]...)
	case cbor.Simple:
		switch {
		case sub < cbor.Float8:
		case sub > cbor.Float64:
			panic(sub)
		default:
			i += 1 << (sub - cbor.Float8)
		}

		return append(w, r[off:i]...)
	case cbor.Labeled:
		w = append(w, r[off:i]...)

		return e.Encode(w, r, i)
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	e.arr = e.JQ.ArrayMap(r, off, e.arr[:0])

	l := len(e.arr)
	if tag == cbor.Map {
		l /= 2
	}

	w = e.CBOR.AppendTag(w, tag, l)

	for _, off := range e.arr {
		w = e.Encode(w, r, off)
	}

	return w
}
