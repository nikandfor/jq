package jqjson

import (
	"encoding/base64"
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
	"nikand.dev/go/json2"
)

type (
	Encoder struct {
		JSON   json2.Emitter
		Base64 *base64.Encoding

		Tag       jq.Tag
		Separator []byte

		arr []jq.Off
		sep bool
	}

	ToOpen  bool
	ToClose bool
	IsNext  bool
)

const (
	Open  ToOpen  = true
	Close ToClose = true
	Next  IsNext  = true
)

func NewEncoder() *Encoder {
	return &Encoder{
		Tag: cbor.String,
	}
}

func (e *Encoder) Reset() {
	e.sep = false
}

func (e *Encoder) ApplyTo(b *jq.Buffer, off jq.Off, next bool) (jq.Off, bool, error) {
	if next {
		return jq.None, false, nil
	}

	var err error

	res := b.Writer().Off()

	tag := e.Tag
	if tag == 0 {
		tag = cbor.String
	}

	var ce cbor.Emitter

	expl := 100
	b.B = ce.AppendTag(b.B, tag, expl)
	st := len(b.B)

	b.B, err = e.Encode(b.B, b, off)
	if err != nil {
		return off, false, err
	}

	b.B = ce.InsertLen(b.B, tag, st, expl, len(b.B)-st)

	return res, false, nil
}

func (e *Encoder) Encode(w []byte, b *jq.Buffer, off jq.Off) (_ []byte, err error) {
	defer func(reset int) {
		if err != nil {
			w = w[:reset]
		}
	}(len(w))

	if e.sep {
		w = append(w, e.Separator...)
	}

	e.sep = true

	return e.encode(w, b, off)
}

func (e *Encoder) encode(w []byte, b *jq.Buffer, off jq.Off) (_ []byte, err error) {
	br := b.Reader()
	tag := br.Tag(off)

	switch tag {
	case cbor.Int, cbor.Neg:
		v := br.Unsigned(off)

		if tag == cbor.Neg {
			w = append(w, '-')
		}

		return strconv.AppendUint(w, v, 10), nil
	case cbor.String:
		return e.encodeString(w, b, off)
	case cbor.Bytes:
		return e.encodeBytes(w, b, off)
	case cbor.Simple:
		sub := br.Simple(off)

		switch sub {
		case cbor.False, cbor.True, cbor.Null:
			lit := []string{"false", "true", "null", "null"}[sub-cbor.False]

			return append(w, lit...), nil
		case cbor.Float8, cbor.Float16, cbor.Float32, cbor.Float64:
			f := br.Float(off)

			return strconv.AppendFloat(w, f, 'f', -1, 64), nil
		case cbor.Undefined, cbor.None:
			return w, jq.NewTypeError(br.TagRaw(off))
		}

		panic(sub)
	case cbor.Label:
		_, _, i := br.Decoder.CBOR.Tag(b.Buf(off))

		return e.encode(w, b, jq.Off(i))
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	w, _, err = e.EncodeContainer(w, b, off, true, true, false)

	return w, err
}

func (e *Encoder) EncodeContainer(w []byte, b *jq.Buffer, off jq.Off, open ToOpen, clos ToClose, next IsNext) (_ []byte, _ IsNext, err error) {
	defer func(reset int) {
		if err != nil {
			w = w[:reset]
		}
	}(len(w))

	br := b.Reader()
	tag := br.Tag(off)

	if tag != cbor.Array && tag != cbor.Map {
		return w, next, jq.NewTypeError(tag, cbor.Array, cbor.Map)
	}

	arrbase := len(e.arr)
	defer func() { e.arr = e.arr[:arrbase] }()

	e.arr = br.ArrayMap(off, e.arr)
	arr := e.arr[arrbase:]

	bt := byte('[')
	if tag == cbor.Map {
		bt = '{'
	}

	w = append(w, bt)

	for j := 0; j < len(arr); j++ {
		if next {
			w = append(w, ',')
		}

		if tag == cbor.Map {
			tag := br.Tag(arr[j])
			if tag != cbor.String {
				return w, next, jq.NewTypeError(tag, cbor.String)
			}

			w, err = e.encode(w, b, arr[j])
			if err != nil {
				return w, next, err
			}

			j++

			w = append(w, ':')
		}

		w, err = e.encode(w, b, arr[j])
		if err != nil {
			return w, next, err
		}

		next = true
	}

	if clos {
		w = append(w, bt+2)
	}

	return w, next, nil
}

func (e *Encoder) encodeString(w []byte, b *jq.Buffer, off jq.Off) ([]byte, error) {
	d := &b.Decoder.CBOR
	r, st := b.Buf(off)

	w = append(w, '"')
	w, i := e.encStr(w, r, st, d)
	w = append(w, '"')
	if i < 0 {
		return w, jq.NewTypeError(jq.Tag(r[st]), cbor.Bytes, cbor.String)
	}

	return w, nil
}

func (e *Encoder) encStr(w, r []byte, i int, d *cbor.Iterator) ([]byte, int) {
	tag, sub, i := d.Tag(r, i)
	l := int(sub)
	if tag != cbor.Bytes && tag != cbor.String {
		return w, -1
	}
	if l >= 0 {
		return e.JSON.AppendStringContent(w, r[i:i+l]), i + l
	}

	for !d.Break(r, &i) {
		w, i = e.encStr(w, r, i, d)
		if i < 0 {
			return w, i
		}
	}

	return w, i
}

func (e *Encoder) encodeBytes(w []byte, b *jq.Buffer, off jq.Off) ([]byte, error) {
	if e.Base64 == nil {
		e.Base64 = base64.StdEncoding
	}

	d := &b.Decoder.CBOR
	r, st := b.Buf(off)

	w = append(w, '"')
	w, i := e.encBytes(w, r, st, d)
	w = append(w, '"')
	if i < 0 {
		return w, jq.NewTypeError(jq.Tag(r[st]), cbor.Bytes, cbor.String)
	}

	return w, nil
}

func (e *Encoder) encBytes(w, r []byte, i int, d *cbor.Iterator) ([]byte, int) {
	tag, sub, i := d.Tag(r, i)
	l := int(sub)
	if tag != cbor.Bytes && tag != cbor.String {
		return w, -1
	}
	if l >= 0 {
		return e.Base64.AppendEncode(w, r[i:i+l]), i + l
	}

	for !d.Break(r, &i) {
		w, i = e.encBytes(w, r, i, d)
		if i < 0 {
			return w, i
		}
	}

	return w, i
}

func (d *Encoder) String() string { return "@json" }
