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

		arr []int
	}
)

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) ApplyTo(b *jq.Buffer, off int, next bool) (int, bool, error) {
	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Bytes && tag != cbor.String {
		return off, false, jq.ErrType
	}

	s := br.Bytes(off)

	res, _, err := d.Decode(b, s, 0)
	if err != nil {
		return off, false, err
	}

	return res, false, nil
}

func (d *Decoder) Decode(b *jq.Buffer, r []byte, st int) (off, i int, err error) {
	bw := b.Writer()

	reset := bw.Len()
	defer bw.ResetIfErr(reset, &err)

	i = st

	tp, i, err := d.JSON.Type(r, st)
	if err != nil {
		return -1, i, err
	}

	switch tp {
	case json.Null, json.Bool:
		c := r[i]

		i, err = d.JSON.Skip(r, i)
		if err != nil {
			return off, st, err
		}

		switch c {
		case 'n':
			return jq.Null, i, nil
		case 't':
			return jq.True, i, nil
		case 'f':
			return jq.False, i, nil
		}

		panic(c)
	case json.Number:
		raw, i, err := d.JSON.Raw(r, i)
		if err != nil {
			return off, st, err
		}

		v := 0
		j := 0

		for j < len(raw) && raw[j] >= '0' && raw[j] <= '9' {
			v = v*10 + int(raw[j]-'0')
			j++
		}

		if j == len(raw) {
			if v == 0 || v == 1 {
				return jq.Zero - v, i, nil
			}

			off = bw.Int(v)

			return off, i, nil
		}

		f, err := strconv.ParseFloat(string(raw), 64)
		if err != nil {
			return off, st, err
		}

		off = bw.Float(f)

		return off, i, nil
	case json.String:
		off = bw.Len()

		n, _, err := d.JSON.DecodedStringLength(r, i)

		b.W = b.Encoder.CBOR.AppendTag(b.W, cbor.String, n)
		b.W, i, err = d.JSON.DecodeString(r, i, b.W)
		if err != nil {
			return off, st, err
		}

		return off, i, nil
	}

	arrbase := len(d.arr)
	defer func() { d.arr = d.arr[:arrbase] }()

	tag := byte(cbor.Array)
	val := 0
	if tp == json.Object {
		tag = cbor.Map
		val = 1
	}

	i, err = d.JSON.Enter(r, i, tp)
	if err != nil {
		return off, st, err
	}

	for d.JSON.ForMore(r, &i, tp, &err) {
		for v := 0; v <= val; v++ {
			off, i, err = d.Decode(b, r, i)
			if err != nil {
				return off, i, err
			}

			d.arr = append(d.arr, off)
		}
	}

	off = bw.ArrayMap(tag, d.arr[arrbase:])

	return off, i, nil
}
