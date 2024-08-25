package jqcbor

import (
	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Encoder struct {
		CBOR cbor.Encoder
		JQ   jq.Decoder

		FilterTag byte

		arr []int
	}
)

func NewEncoder() *Encoder {
	return &Encoder{
		FilterTag: cbor.String,
	}
}

func (e *Encoder) ApplyTo(b *jq.Buffer, off int, next bool) (int, bool, error) {
	var err error

	res := b.Writer().Len()
	r0, r1 := b.Unwrap()

	tag := e.FilterTag
	if tag == 0 {
		tag = cbor.String
	}

	var ce cbor.Encoder

	b.W = append(b.W, 0)
	st := len(b.W)

	b.W, err = e.Encode(b.W, r0, r1, off)
	if err != nil {
		return off, false, err
	}

	b.W = ce.InsertLen(b.W, tag, st, len(b.W)-st)

	return res, false, nil
}

func (e *Encoder) Encode(w, r0, r1 []byte, off int) ([]byte, error) {
	var err error
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

		return append(w, buf[off:i]...), nil
	case cbor.Simple:
		switch {
		case sub < cbor.Float8:
		case sub <= cbor.Float64:
			i += 1 << (sub - cbor.Float8)
		default:
			panic(sub)
		}

		return append(w, buf[off:i]...), nil
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
		w, err = e.Encode(w, r0, r1, off)
		if err != nil {
			return w, err
		}
	}

	return w, nil
}
