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

func (e *Encoder) Encode(w, r0, r1 []byte, off int) []byte {
	var buf []byte
	var base int

	if off < len(r0) {
		buf, base, off = r0, 0, off
	} else {
		buf, base, off = r1, len(r0), off-len(r0)
	}

	tag, sub, i := e.JQ.CBOR.Tag(buf, off)

	switch tag {
	case cbor.Int, cbor.Neg, cbor.Bytes, cbor.String:
		if tag == cbor.Bytes || tag == cbor.String {
			i += int(sub)
		}

		return append(w, buf[off:i]...)
	case cbor.Simple:
		switch {
		case sub < cbor.Float8:
		case sub <= cbor.Float64:
			i += 1 << (sub - cbor.Float8)
		default:
			panic(sub)
		}

		return append(w, buf[off:i]...)
	case cbor.Labeled:
		w = append(w, buf[off:i]...)

		return e.Encode(w, r0, r1, i)
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	arrbase := len(e.arr)
	defer func() { e.arr = e.arr[:arrbase] }()

	e.arr, _ = e.JQ.ArrayMap(buf, base, off, e.arr)

	l := len(e.arr)
	if tag == cbor.Map {
		l /= 2
	}

	w = e.CBOR.AppendTag(w, tag, l)

	for _, off := range e.arr[arrbase:] {
		w = e.Encode(w, r0, r1, off)
	}

	return w
}
