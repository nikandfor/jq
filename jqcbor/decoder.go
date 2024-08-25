package jqcbor

import (
	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Decoder struct {
		CBOR cbor.Decoder
		JQ   jq.Encoder

		arr []int
	}
)

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) ApplyTo(b *jq.Buffer, off int, next bool) (int, bool, error) {
	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.String && tag != cbor.Bytes {
		return off, false, jq.ErrType
	}

	s := br.Bytes(off)

	w, res, _, err := d.Decode(b.W, s, len(b.R), 0)
	if err != nil {
		return off, false, err
	}

	b.W = w

	return res, false, nil
}

func (d *Decoder) Decode(w, r []byte, base, st int) (_ []byte, off, i int, err error) {
	defer func(l int) {
		if err == nil {
			return
		}

		w = w[:l]
	}(len(w))

	i = st
	off = base + len(w)

	tag, sub, i := d.CBOR.Tag(r, i)

	switch tag {
	case cbor.Int, cbor.Neg, cbor.Bytes, cbor.String:
		if tag == cbor.Bytes || tag == cbor.String {
			i += int(sub)
		}

		w = append(w, r[st:i]...)

		return w, off, i, nil
	case cbor.Simple:
		switch {
		case sub < cbor.Float8:
		case sub > cbor.Float64:
			panic(sub)
		default:
			i += 1 << (sub - cbor.Float8)
		}

		w = append(w, r[st:i]...)

		return w, off, i, nil
	case cbor.Labeled:
		w = append(w, r[st:i]...)

		return d.Decode(w, r, base, i)
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	arrbase := len(d.arr)
	defer func() { d.arr = d.arr[:arrbase] }()

	val := 0
	if tag == cbor.Map {
		val = 1
	}

	for j := 0; sub < 0 && !d.CBOR.Break(r, &i) || j < int(sub); j++ {
		for v := 0; v <= val; v++ {
			w, off, i, err = d.Decode(w, r, base, i)
			if err != nil {
				return
			}

			d.arr = append(d.arr, off)
		}
	}

	off = base + len(w)
	w = d.JQ.AppendArrayMap(w, tag, off, d.arr[arrbase:])

	return w, off, i, nil
}
