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
	b.W = ce.AppendTag(b.W, tag, expl)
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
	case cbor.Int, cbor.Neg:
		v := br.Unsigned(off)

		if tag == cbor.Neg {
			w = append(w, '-')
		}

		return strconv.AppendUint(w, v, 10), nil
	case cbor.String:
		v := br.Bytes(off)

		return e.JSON.AppendString(w, v), nil
	case cbor.Bytes:
		if e.Base64 == nil {
			e.Base64 = base64.StdEncoding
		}

		v := br.Bytes(off)

		w = append(w, '"')
		w = e.Base64.AppendEncode(w, v)
		w = append(w, '"')

		return w, nil
	case cbor.Simple:
		sub := br.Simple(off)

		switch sub {
		case cbor.False, cbor.True, cbor.Null:
			lit := []string{"false", "true", "null", "null"}[tag&cbor.SubMask-cbor.False]

			return append(w, lit...), nil
		case cbor.Float8, cbor.Float16, cbor.Float32, cbor.Float64:
			f := br.Float(off)

			return strconv.AppendFloat(w, f, 'v', -1, 64), nil
		case cbor.Undefined, cbor.None:
			return w, jq.ErrType
		default:
			panic(sub)
		}
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

	open := byte('[')
	if tag == cbor.Map {
		open = '{'
	}

	w = append(w, open)

	for j := arrbase; j < len(e.arr); j++ {
		if j != arrbase {
			w = append(w, ',')
		}

		if tag == cbor.Map {
			tag := br.Tag(e.arr[j])
			if tag != cbor.String {
				return w, jq.ErrType
			}

			w, err = e.Encode(w, b, e.arr[j])
			if err != nil {
				return w, err
			}

			j++

			w = append(w, ':')
		}

		w, err = e.Encode(w, b, e.arr[j])
		if err != nil {
			return w, err
		}
	}

	w = append(w, open+2)

	return w, nil
}
