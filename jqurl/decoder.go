package jqurl

import (
	"maps"
	"net/url"
	"slices"
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Decoder struct {
		arr []jq.Off
	}
)

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

	res, _, err := d.Decode(b, s, 0)
	if err != nil {
		return off, false, err
	}

	return res, false, nil
}

func (d *Decoder) Decode(b *jq.Buffer, r []byte, st int) (off jq.Off, i int, err error) {
	vals, err := url.ParseQuery(string(r[st:]))
	if err != nil {
		return jq.None, st, err
	}

	bw := b.Writer()

	reset := bw.Off()
	defer bw.ResetIfErr(reset, &err)

	d.arr = d.arr[:0]

	for _, k := range slices.Sorted(maps.Keys(vals)) {
		vs := vals[k]

		key := bw.String(k)
		val := d.decode(b, vs)

		d.arr = append(d.arr, key, val)
	}

	return bw.Map(d.arr), len(r), nil
}

func (d *Decoder) decode(b *jq.Buffer, vs []string) jq.Off {
	if len(vs) == 0 || len(vs) == 1 && vs[0] == "" {
		return jq.Null
	}
	if len(vs) == 1 {
		return d.decodeOne(b, vs[0])
	}

	reset := len(d.arr)
	defer func() { d.arr = d.arr[:reset] }()

	for _, v := range vs {
		d.arr = append(d.arr, d.decodeOne(b, v))
	}

	return b.Writer().Array(d.arr[reset:])
}

func (d *Decoder) decodeOne(b *jq.Buffer, v string) jq.Off {
	switch v {
	case "":
		return jq.EmptyString
	case "0":
		return jq.Zero
	case "1":
		return jq.One
	case "false":
		return jq.False
	case "true":
		return jq.True
	}

	bw := b.Writer()

	x, err := strconv.ParseUint(v, 10, 64)
	if err == nil {
		return bw.Uint64(x)
	}

	y, err := strconv.ParseInt(v, 10, 64)
	if err == nil {
		return bw.Int64(y)
	}

	z, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return bw.Float(z)
	}

	return bw.String(v)
}
