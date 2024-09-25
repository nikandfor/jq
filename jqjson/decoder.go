package jqjson

import (
	"errors"
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
	"nikand.dev/go/json"
)

type (
	Decoder struct {
		JSON json.Decoder

		arr []jq.Off
	}

	MultiDecoder struct {
		*Decoder
		i int
	}
)

var ErrPartialRead = errors.New("partial read")

func NewMultiDecoder() *MultiDecoder {
	return &MultiDecoder{
		Decoder: NewDecoder(),
	}
}

func (d *MultiDecoder) ApplyTo(b *jq.Buffer, off jq.Off, next bool) (jq.Off, bool, error) {
	if !next {
		d.i = 0
	}

	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Bytes && tag != cbor.String {
		return off, false, jq.NewTypeError(tag, cbor.Bytes, cbor.String)
	}

	s := br.Bytes(off)

	d.i = d.JSON.SkipSpaces(s, d.i)
	if d.i == len(s) {
		return jq.None, false, nil
	}

	res, i, err := d.decode(b, s, d.i, false)
	if err != nil {
		return off, false, err
	}

	d.i = d.JSON.SkipSpaces(s, i)

	return res, d.i < len(s), nil
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) ApplyTo(b *jq.Buffer, off jq.Off, next bool) (jq.Off, bool, error) {
	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Bytes && tag != cbor.String {
		return off, false, jq.NewTypeError(tag, cbor.Bytes, cbor.String)
	}

	s := br.Bytes(off)

	res, i, err := d.Decode(b, s, 0)
	if err != nil {
		return off, false, err
	}

	i = d.JSON.SkipSpaces(s, i)
	if i != len(s) {
		return jq.None, false, ErrPartialRead
	}

	return res, false, nil
}

func (d *Decoder) Decode(b *jq.Buffer, r []byte, st int) (off jq.Off, i int, err error) {
	return d.decode(b, r, st, false)
}

func (d *Decoder) decode(b *jq.Buffer, r []byte, st int, key bool) (off jq.Off, i int, err error) {
	bw := b.Writer()

	reset := bw.Off()
	defer bw.ResetIfErr(reset, &err)

	i = d.JSON.SkipSpaces(r, st)
	if i == len(r) {
		return jq.None, i, nil
	}

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
				return jq.Zero - jq.Off(v), i, nil
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
		off = bw.Off()

		n, _, err := d.JSON.DecodedStringLength(r, i)

		b.B = b.Encoder.CBOR.AppendTag(b.B, cbor.String, n)
		b.B, i, err = d.JSON.DecodeString(r, i, b.B)
		if err != nil {
			return off, st, err
		}

		return off, i, nil
	}

	arrbase := len(d.arr)
	defer func() { d.arr = d.arr[:arrbase] }()

	tag := cbor.Array
	if tp == json.Object {
		tag = cbor.Map
	}

	i, err = d.JSON.Enter(r, i, tp)
	if err != nil {
		return off, st, err
	}

	for d.JSON.ForMore(r, &i, tp, &err) {
		if tp == json.Object {
			off, i, err = d.decode(b, r, i, true)
			if err != nil {
				return off, i, err
			}

			d.arr = append(d.arr, off)
		}

		off, i, err = d.decode(b, r, i, false)
		if err != nil {
			return off, i, err
		}

		d.arr = append(d.arr, off)
	}

	off = bw.ArrayMap(tag, d.arr[arrbase:])

	return off, i, nil
}

func (d *Decoder) String() string { return "fromjson" }
