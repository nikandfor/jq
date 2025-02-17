package jqcbor

import (
	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Encoder struct {
		CBOR cbor.Emitter

		Tag jq.Tag

		arr []Off
	}
)

func NewEncoder() *Encoder {
	return &Encoder{
		CBOR: cbor.Emitter{Flags: cbor.FtDefault},
		Tag:  cbor.String,
	}
}

func (e *Encoder) ApplyTo(b *jq.Buffer, off Off, next bool) (Off, bool, error) {
	var err error

	res := b.Writer().Off()

	tag := e.Tag
	if tag == 0 {
		tag = cbor.String
	}

	var ce cbor.Emitter

	expl := 100
	b.B = ce.AppendTag(b.B, tag, 100)
	st := len(b.B)

	b.B, err = e.Encode(b.B, b, off)
	if err != nil {
		return off, false, err
	}

	b.B = ce.InsertLen(b.B, tag, st, expl, len(b.B)-st)

	return res, false, nil
}

func (e *Encoder) Encode(w []byte, b *jq.Buffer, off Off) (_ []byte, err error) {
	br := b.Reader()

	tag := br.Tag(off)

	switch tag {
	case cbor.Int, cbor.Neg, cbor.Bytes, cbor.String, cbor.Simple:
		raw := br.Raw(off)

		return append(w, raw...), nil
	case cbor.Label:
		lab := br.Label(off)
		if lab < jq.LabelOffset {
			panic(lab)
		}

		w = e.CBOR.AppendLabel(w, lab-jq.LabelOffset)

		return e.Encode(w, b, br.UnderLabel(off))
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	arrbase := len(e.arr)
	defer func() { e.arr = e.arr[:arrbase] }()

	e.arr = br.ArrayMap(off, e.arr)

	l := len(e.arr[arrbase:])
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

func (e *Encoder) String() string { return "@cbor" }
