package jqjson

import (
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
	"nikand.dev/go/json"
)

type (
	Decoder struct {
		JSON json.Decoder
		JQ   jq.Encoder

		arr []int
	}
)

func (d *Decoder) ApplyTo(b *jq.Buffer, off int, next bool) (int, bool, error) {
	br := b.Reader()
	tag := br.Tag(off)
	if tag != cbor.Bytes && tag != cbor.String {
		return off, false, jq.ErrType
	}

	s := br.Bytes(off)

	w, off, _, err := d.Decode(b.W, s, len(b.R), 0)
	if err != nil {
		return off, false, err
	}

	b.W = w

	return off, false, nil
}

func (d *Decoder) Decode(w, r []byte, base, st int) (_ []byte, off, i int, err error) {
	defer func(l int) {
		if err == nil {
			return
		}

		w = w[:l]
	}(len(w))

	var raw []byte

	i = st
	off = len(w)

	tp, i, err := d.JSON.Type(r, st)
	if err != nil {
		return w, -1, i, err
	}

	switch tp {
	case json.Null, json.Bool:
		c := r[i]

		i, err = d.JSON.Skip(r, i)
		if err != nil {
			return w, off, st, err
		}

		switch c {
		case 'n':
			w = d.JQ.AppendNull(w)
		case 't', 'f':
			w = d.JQ.AppendBool(w, c == 't')
		}

		return w, off, i, nil
	case json.Number:
		raw, i, err = d.JSON.Raw(r, i)
		if err != nil {
			return
		}

		v := 0
		j := 0

		for j < len(raw) && raw[j] >= '0' && raw[j] <= '9' {
			v = v*10 + int(raw[j]-'0')
			j++
		}

		if j == len(raw) {
			w = d.JQ.AppendInt(w, v)

			return w, off, i, nil
		}

		f, err := strconv.ParseFloat(string(raw), 64)
		if err != nil {
			return w, off, st, err
		}

		w = d.JQ.AppendFloat(w, f)

		return w, off, i, nil
	case json.String:
		w = d.JQ.CBOR.AppendTag(w, cbor.String, 0)

		st := len(w)
		w, i, err = d.JSON.DecodeString(r, i, w)
		if err != nil {
			return w, off, st, err
		}

		w = d.JQ.CBOR.InsertLen(w, cbor.String, st, len(w)-st)

		return w, off, i, nil
	}

	tag := byte(cbor.Array)
	val := 0
	if tp == json.Object {
		tag = cbor.Map
		val = 1
	}

	arrbase := len(d.arr)
	defer func() { d.arr = d.arr[:arrbase] }()

	i, err = d.JSON.Enter(r, i, tp)
	if err != nil {
		return w, off, st, err
	}

	for d.JSON.ForMore(r, &i, tp, &err) {
		for v := 0; v <= val; v++ {
			w, off, i, err = d.Decode(w, r, base, i)
			if err != nil {
				return w, off, i, err
			}

			d.arr = append(d.arr, base+off)
		}
	}

	off = len(w)
	w = d.JQ.AppendArrayMap(w, tag, off, d.arr[arrbase:])

	return w, off, i, nil
}
