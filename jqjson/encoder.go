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

		arr []int
	}
)

func (e *Encoder) ApplyTo(b *jq.Buffer, off int, next bool) (int, bool, error) {
	res := b.Writer().Len()

	r0, r1 := b.Unwrap()
	b.W = e.Encode(b.W, r0, r1, off)

	return res, false, nil
}

func (e *Encoder) Encode(w, r0, r1 []byte, off int) []byte {
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

		return strconv.AppendUint(w, v, 10)
	case cbor.String:
		v, _ := e.JQ.CBOR.Bytes(buf, off)

		return e.JSON.AppendString(w, v)
	case cbor.Bytes:
		if e.Base64 == nil {
			e.Base64 = base64.StdEncoding
		}

		v, _ := e.JQ.CBOR.Bytes(buf, off)

		w = append(w, '"')
		w = e.Base64.AppendEncode(w, v)
		w = append(w, '"')

		return w
	case cbor.Simple:
		_, sub, _ := e.JQ.CBOR.Tag(buf, off)

		switch sub {
		case cbor.False, cbor.True, cbor.Null, cbor.Undefined:
			lit := []string{"false", "true", "null", "null"}[tag&cbor.SubMask-cbor.False]

			return append(w, lit...)
		case cbor.Float8, cbor.Float16, cbor.Float32, cbor.Float64:
			f, _ := e.JQ.CBOR.Float(buf, off)

			return strconv.AppendFloat(w, f, 'v', -1, 64)
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
			w = e.Encode(w, r0, r1, e.arr[j])
			j++

			w = append(w, ':')
		}

		w = e.Encode(w, r0, r1, e.arr[j])
	}

	w = append(w, br+2)

	return w
}
