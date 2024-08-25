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
		JQ     jq.Decoder
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

	tag := e.JQ.TagOnly(buf, off)

	switch tag {
	case cbor.Int, cbor.Neg:
		v, _ := e.JQ.CBOR.Unsigned(buf, off)

		if tag == cbor.Neg {
			w = append(w, '-')
		}

		return strconv.AppendUint(w, v, 10), nil
	case cbor.String:
		v, _ := e.JQ.CBOR.Bytes(buf, off)

		return e.JSON.AppendString(w, v), nil
	case cbor.Bytes:
		if e.Base64 == nil {
			e.Base64 = base64.StdEncoding
		}

		v, _ := e.JQ.CBOR.Bytes(buf, off)

		w = append(w, '"')
		w = e.Base64.AppendEncode(w, v)
		w = append(w, '"')

		return w, nil
	case cbor.Simple:
		_, sub, _ := e.JQ.CBOR.Tag(buf, off)

		switch sub {
		case cbor.False, cbor.True, cbor.Null:
			lit := []string{"false", "true", "null", "null"}[tag&cbor.SubMask-cbor.False]

			return append(w, lit...), nil
		case cbor.Float8, cbor.Float16, cbor.Float32, cbor.Float64:
			f, _ := e.JQ.CBOR.Float(buf, off)

			return strconv.AppendFloat(w, f, 'v', -1, 64), nil
		case cbor.Undefined, cbor.None:
			return w, jq.ErrType
		default:
			panic(sub)
		}
	case cbor.Labeled:
		_, _, i := e.JQ.CBOR.Tag(buf, off)

		return e.Encode(w, r0, r1, i)
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	arrbase := len(e.arr)
	defer func() { e.arr = e.arr[:arrbase] }()

	e.arr, _ = e.JQ.ArrayMap(buf, base, off, e.arr)

	br := byte('[')
	if tag == cbor.Map {
		br = '{'
	}

	w = append(w, br)

	for j := arrbase; j < len(e.arr); j++ {
		if j != arrbase {
			w = append(w, ',')
		}

		if tag == cbor.Map {
			w, err = e.Encode(w, r0, r1, e.arr[j])
			if err != nil {
				return w, err
			}

			j++

			w = append(w, ':')
		}

		w, err = e.Encode(w, r0, r1, e.arr[j])
		if err != nil {
			return w, err
		}
	}

	w = append(w, br+2)

	return w, nil
}
