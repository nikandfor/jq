package jqjson

import (
	"encoding/base64"
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
	"nikand.dev/go/json"
)

type (
	Encoder struct {
		JSON   json.Encoder
		Base64 *base64.Encoding

		FilterTag byte
		Separator []byte

		arr []Off
		sep bool
	}

	RawEncoder struct {
		*Encoder
	}
)

func NewRawEncoder() *RawEncoder {
	return &RawEncoder{Encoder: NewEncoder()}
}

func (e *RawEncoder) ApplyTo(b *jq.Buffer, off Off, next bool) (Off, bool, error) {
	if !next {
		e.sep = false
	}

	var err error
	res := b.Writer().Off()

	b.W, err = e.Encode(b.W, b, off)
	if err != nil {
		return off, false, err
	}

	return res, false, nil
}

func (e *RawEncoder) Encode(w []byte, b *jq.Buffer, off Off) (_ []byte, err error) {
	if off == jq.None {
		return w, nil
	}

	w, err = e.Encoder.Encode(w, b, off)
	if err != nil {
		return w, err
	}

	return w, nil
}

func NewEncoder() *Encoder {
	return &Encoder{
		FilterTag: cbor.String,
	}
}

func (e *Encoder) Reset() {
	e.sep = false
}

func (e *Encoder) ApplyTo(b *jq.Buffer, off Off, next bool) (Off, bool, error) {
	if next {
		return jq.None, false, nil
	}

	var err error

	res := b.Writer().Off()

	tag := e.FilterTag
	if tag == 0 {
		tag = cbor.String
	}

	var ce cbor.Encoder

	expl := 100
	b.W = ce.AppendTag(b.W, tag, expl)
	st := len(b.W)

	b.W, err = e.Encode(b.W, b, off)
	if err != nil {
		return off, false, err
	}

	b.W = ce.InsertLen(b.W, tag, st, expl, len(b.W)-st)

	return res, false, nil
}

func (e *Encoder) Encode(w []byte, b *jq.Buffer, off Off) (_ []byte, err error) {
	defer func(reset int) {
		if err != nil {
			w = w[:reset]
		}
	}(len(w))

	if e.sep {
		w = append(w, e.Separator...)
	}

	e.sep = true

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
		default:
			panic(sub)
		}
	case cbor.Labeled:
		_, _, i := br.Decoder.CBOR.Tag(b.Buf(off))

		return e.Encode(w, b, Off(i))
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	w, _, err = e.EncodeContainer(w, b, off, true, true, false)
	return w, err
}

func (e *Encoder) EncodeContainer(w []byte, b *jq.Buffer, off Off, open, clos, next bool) (_ []byte, _ bool, err error) {
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

			w, err = e.Encode(w, b, arr[j])
			if err != nil {
				return w, next, err
			}

			j++

			w = append(w, ':')
		}

		w, err = e.Encode(w, b, arr[j])
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

func (e *Encoder) encodeString(w []byte, b *jq.Buffer, off Off) ([]byte, error) {
	d := &b.Decoder.CBOR
	r, st := b.Buf(off)

	w = append(w, '"')
	w, i := e.encStr(w, r, st, d)
	w = append(w, '"')
	if i < 0 {
		return w, jq.NewTypeError(r[st], cbor.Bytes, cbor.String)
	}

	return w, nil
}

func (e *Encoder) encStr(w, r []byte, i int, d *cbor.Decoder) ([]byte, int) {
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

func (e *Encoder) encodeBytes(w []byte, b *jq.Buffer, off Off) ([]byte, error) {
	if e.Base64 == nil {
		e.Base64 = base64.StdEncoding
	}

	d := &b.Decoder.CBOR
	r, st := b.Buf(off)

	w = append(w, '"')
	w, i := e.encBytes(w, r, st, d)
	w = append(w, '"')
	if i < 0 {
		return w, jq.NewTypeError(r[st], cbor.Bytes, cbor.String)
	}

	return w, nil
}

func (e *Encoder) encBytes(w, r []byte, i int, d *cbor.Decoder) ([]byte, int) {
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
