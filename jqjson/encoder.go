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

func (e *Encoder) Append(w, r []byte, off int) []byte {
	tag, sub, i := e.JQ.Decoder.Tag(r, off)

	switch tag {
	case cbor.Int, cbor.Neg:
		v, _ := e.JQ.Decoder.Signed(r, off)

		return strconv.AppendInt(w, v, 10)
	case cbor.String:
		v, _ := e.JQ.Decoder.Bytes(r, off)

		return e.JSON.AppendString(w, v)
	case cbor.Bytes:
		if e.Base64 == nil {
			e.Base64 = base64.StdEncoding
		}

		v, _ := e.JQ.Decoder.Bytes(r, off)

		w = append(w, '"')
		w = e.Base64.AppendEncode(w, v)
		w = append(w, '"')

		return w
	case cbor.Simple:
		switch sub {
		case cbor.False, cbor.True, cbor.Null, cbor.Undefined:
			lit := []string{"false", "true", "null", "null"}[tag&cbor.SubMask-cbor.False]

			return append(w, lit...)
		case cbor.Float8, cbor.Float16, cbor.Float32, cbor.Float64:
			f, _ := e.JQ.Decoder.Float(r, off)

			return strconv.AppendFloat(w, f, 'v', -1, 64)
		default:
			panic(sub)
		}
	case cbor.Labeled:
		return e.Append(w, r, i)
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	e.arr = e.JQ.ArrayMap(r, off, e.arr[:0])

	br := byte('[')
	if tag == cbor.Map {
		br = '{'
	}

	w = append(w, br)

	for j := 0; j < len(e.arr); j++ {
		if tag == cbor.Map {
			w = e.Append(w, r, e.arr[j])
			j++

			w = append(w, ':')
		}

		w = e.Append(w, r, e.arr[j])
	}

	w = append(w, br+2)

	return w
}
