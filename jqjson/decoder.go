package jqjson

import (
	"errors"
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
	"nikand.dev/go/json"
)

type (
	Off = jq.Off

	Decoder struct {
		JSON json.Decoder

		DeduplicateKeys bool

		arr  []Off
		keys map[string]Off
	}

	RawDecoder struct {
		*Decoder
		i int
	}

	MultiDecoder struct {
		*Decoder
		i int
	}
)

var ErrPartialRead = errors.New("partial read")

func NewRawDecoder() *RawDecoder {
	return &RawDecoder{
		Decoder: NewDecoder(),
	}
}

func (d *RawDecoder) ApplyTo(b *jq.Buffer, off Off, next bool) (Off, bool, error) {
	if !next {
		d.i = 0

		if d.DeduplicateKeys && d.keys == nil {
			d.keys = map[string]Off{}
		}

		clear(d.keys)
	}

	res, i, err := d.Decode(b, b.R, d.i)
	if err != nil {
		return off, false, err
	}

	d.i = i

	return res, i < len(b.R), nil
}

func (d *RawDecoder) DecodeAll(b *jq.Buffer, r []byte, st int, arr []Off) (_ []Off, i int, err error) {
	var off Off
	i = st

	for i < len(r) {
		off, i, err = d.Decode(b, r, i)
		if err != nil {
			return arr, i, err
		}

		if off != jq.None {
			arr = append(arr, off)
		}
	}

	return arr, i, nil
}

func (d *RawDecoder) Decode(b *jq.Buffer, r []byte, st int) (Off, int, error) {
	i := d.JSON.SkipSpaces(r, st)
	if i >= len(r) {
		return jq.None, i, nil
	}

	off, i, err := d.decode(b, r, i, false)
	if err != nil {
		return jq.None, i, err
	}

	i = d.JSON.SkipSpaces(r, i)

	return off, i, nil
}

func NewMultiDecoder() *MultiDecoder {
	return &MultiDecoder{
		Decoder: NewDecoder(),
	}
}

func (d *MultiDecoder) ApplyTo(b *jq.Buffer, off Off, next bool) (Off, bool, error) {
	if !next {
		d.i = 0

		if d.DeduplicateKeys && d.keys == nil {
			d.keys = map[string]Off{}
		}

		clear(d.keys)
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

func (d *Decoder) ApplyTo(b *jq.Buffer, off Off, next bool) (Off, bool, error) {
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

func (d *Decoder) Decode(b *jq.Buffer, r []byte, st int) (off Off, i int, err error) {
	if d.DeduplicateKeys {
		if d.keys == nil {
			d.keys = map[string]Off{}
		}
	}

	defer clear(d.keys)

	return d.decode(b, r, st, false)
}

func (d *Decoder) decode(b *jq.Buffer, r []byte, st int, key bool) (off Off, i int, err error) {
	bw := b.Writer()

	reset := bw.Off()
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
				return jq.Zero - Off(v), i, nil
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
		reset := len(b.W)
		off = bw.Off()

		n, _, err := d.JSON.DecodedStringLength(r, i)

		b.W = b.Encoder.CBOR.AppendTag(b.W, cbor.String, n)
		b.W, i, err = d.JSON.DecodeString(r, i, b.W)
		if err != nil {
			return off, st, err
		}

		if off, ok := d.keys[string(b.W[reset:])]; ok {
			b.W = b.W[:reset]

			return off, i, nil
		}

		if d.DeduplicateKeys {
			d.keys[string(b.W[reset:])] = off
		}

		return off, i, nil
	}

	arrbase := len(d.arr)
	defer func() { d.arr = d.arr[:arrbase] }()

	tag := byte(cbor.Array)
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
