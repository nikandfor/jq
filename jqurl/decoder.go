package jqurl

import (
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
	"nikand.dev/go/skip"
)

type (
	Decoder struct {
		Fold bool

		arr []jq.Off
		b   []byte
	}
)

func NewDecoder() *Decoder {
	return &Decoder{
		Fold: true,
	}
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
	i = st
	bw := b.Writer()

	reset := bw.Off()
	defer bw.ResetIfErr(reset, &err)

	d.arr = d.arr[:0]
	d.b = d.b[:0]

	var s skip.Str

	for i < len(r) {
		br := len(d.b)

		//	log.Printf("skip url %v %v %q", i, len(r), r[i])

		s, d.b, _, i = skip.DecodeURLQuery(r, i, 0, d.b)
		if s.Err() {
			return jq.None, i, s
		}

		key := bw.TagBytes(cbor.String, d.b[br:])
		val := jq.Null
		br = len(d.b)

		if i < len(r) && r[i] == '=' {
			i++

			s, d.b, _, i = skip.DecodeURLQuery(r, i, skip.URLValue, d.b)
			if s.Err() {
				return jq.None, i, s
			}

			val = d.decodeVal(b, string(d.b[br:]))
		}

		if i < len(r) && r[i] == '&' {
			i++
		}

		d.arr = append(d.arr, key, val)
	}

	if !d.Fold {
		return bw.Map(d.arr), i, nil
	}

	for j := 0; j < len(d.arr); j += 2 {
		l := len(d.arr)
		d.arr = append(d.arr, d.arr[j+1])

		w := j + 2

		for r := w; r < l; r += 2 {
			if d.arr[r] == d.arr[j] {
				d.arr = append(d.arr, d.arr[r+1])
				continue
			}

			d.arr[w], d.arr[w+1] = d.arr[r], d.arr[r+1]
			w += 2
		}

		if w != l {
			d.arr[j+1] = bw.Array(d.arr[l:])
		}

		d.arr = d.arr[:w]
	}

	return bw.Map(d.arr), i, nil
}

func (d *Decoder) decodeVal(b *jq.Buffer, v string) jq.Off {
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

func (d *Decoder) String() string { return "@urid" }
