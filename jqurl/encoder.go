package jqurl

import (
	"net/url"
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Encoder struct {
		Tag jq.Tag

		arr []jq.Off
	}
)

func NewEncoder() *Encoder { return &Encoder{Tag: cbor.String} }

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

	return e.encode(w, b, off, 0)
}

func (e *Encoder) encode(w []byte, b *jq.Buffer, off jq.Off, depth int) (_ []byte, err error) {
	br := b.Reader()
	tag := br.Tag(off)

	switch tag {
	case cbor.Int, cbor.Neg:
		v := br.Unsigned(off)

		if tag == cbor.Neg {
			w = append(w, '-')
		}

		return strconv.AppendUint(w, v, 10), nil
	case cbor.String, cbor.Bytes:
		b := br.Bytes(off)
		e := url.QueryEscape(string(b))

		return append(w, e...), nil
	case cbor.Simple:
		sub := br.Simple(off)

		switch sub {
		case cbor.Null:
			return w, nil
		case cbor.False, cbor.True:
			lit := []string{"false", "true", "null", "null"}[sub-cbor.False]

			return append(w, lit...), nil
		case cbor.Float8, cbor.Float16, cbor.Float32, cbor.Float64:
			f := br.Float(off)

			return strconv.AppendFloat(w, f, 'f', -1, 64), nil
		case cbor.Undefined, cbor.None:
			return w, jq.NewTypeError(br.TagRaw(off))
		}

		panic(sub)
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	if depth > 0 {
		return w, jq.NewTypeError(tag)
	}

	arrbase := len(e.arr)
	defer func() { e.arr = e.arr[:arrbase] }()

	e.arr = br.ArrayMap(off, e.arr)
	arr := e.arr[arrbase:]

	for j := 0; j < len(arr); {
		if j != 0 {
			w = append(w, '&')
		}

		if tag == cbor.Array {
			w, err = e.encode(w, b, arr[j], depth+1)
			if err != nil {
				return w, err
			}

			j++

			continue
		}

		w, err = e.encodeMapPair(w, b, arr[j], arr[j+1], depth+1)
		if err != nil {
			return w, err
		}

		j += 2
	}

	return w, nil
}

func (e *Encoder) encodeMapPair(w []byte, b *jq.Buffer, key, val jq.Off, depth int) (_ []byte, err error) {
	br := b.Reader()
	tag := br.Tag(val)

	if tag == cbor.Map {
		return w, jq.NewTypeError(tag)
	}
	if tag != cbor.Array {
		return e.encodePair(w, b, key, val, depth)
	}

	arrbase := len(e.arr)
	defer func() { e.arr = e.arr[:arrbase] }()

	e.arr = br.ArrayMap(val, e.arr)
	arr := e.arr[arrbase:]

	for j := 0; j < len(arr); j++ {
		if j != 0 {
			w = append(w, '&')
		}

		w, err = e.encodePair(w, b, key, arr[j], depth)
		if err != nil {
			return w, err
		}
	}

	return w, nil
}

func (e *Encoder) encodePair(w []byte, b *jq.Buffer, key, val jq.Off, depth int) (_ []byte, err error) {
	w, err = e.encode(w, b, key, depth)
	if err != nil {
		return w, err
	}

	if b.Equal(val, jq.Null) {
		return w, nil
	}

	w = append(w, '=')

	w, err = e.encode(w, b, val, depth)
	if err != nil {
		return w, err
	}

	return w, nil
}

func (e *Encoder) String() string { return "@uri" }
