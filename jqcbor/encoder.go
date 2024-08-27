package jqcbor

import (
	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Encoder struct {
		CBOR cbor.Encoder

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

	tag := e.FilterTag
	if tag == 0 {
		tag = cbor.String
	}

	var ce cbor.Encoder

	expl := 100
	b.W = ce.AppendTag(b.W, tag, 100)
	st := len(b.W)

	b.W, err = e.Encode(b.W, b, off)
	if err != nil {
		return off, false, err
	}

	b.W = ce.InsertLen(b.W, tag, st, expl, len(b.W)-st)

	return res, false, nil
}

func (e *Encoder) Encode(w []byte, b *jq.Buffer, off int) (_ []byte, err error) {
	br := b.Reader()

	tag := br.Tag(off)

	switch tag {
	case cbor.Int, cbor.Neg, cbor.Bytes, cbor.String, cbor.Simple:
		raw := br.Raw(off)

		return append(w, raw...), nil
	case cbor.Labeled:
		_, _, i := br.Decoder.CBOR.Tag(b.Buf(off))

		return e.Encode(w, b, i)
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	arrbase := len(e.arr)
	defer func() { e.arr = e.arr[:arrbase] }()

	e.arr = br.ArrayMap(off, e.arr)

	l := len(e.arr)
	if tag == cbor.Map {
		l /= 2
	}

	w = e.CBOR.AppendTag(w, tag, l)

	for _, off := range e.arr[arrbase:] {
		w, err = e.Encode(w, b, off)
		if err != nil {
			return w, err
		}
	}

	return w, nil
}
