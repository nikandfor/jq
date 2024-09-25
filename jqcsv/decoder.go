package jqcsv

import (
	"fmt"
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Decoder struct {
		Tag    jq.Tag
		Comma  byte
		Header bool

		header bool

		arr []jq.Off
	}
)

func NewDecoder() *Decoder {
	return &Decoder{
		Tag:   cbor.Array,
		Comma: ',',
	}
}

func (d *Decoder) Reset() {
	d.header = false
	d.arr = d.arr[:0]
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
	comma := csel(d.Comma != 0, d.Comma, ',')
	tag := csel(d.Tag != 0, d.Tag, cbor.Array)
	val := csel(tag == cbor.Map, 1, 0)
	i = st

	for i < len(r) && (r[i] == '\n' || r[i] == '\r') {
		i++
	}
	if i == len(r) {
		return jq.None, i, nil
	}

	for {
		col := 0
		hdr := csel(d.Header && tag == cbor.Map && !d.header, 1, 0)

		for i < len(r) {
			if r[i] == '\n' || r[i] == '\r' {
				break
			}

			off, i, err = d.decodeOne(b, r, i)
			if err != nil {
				return jq.None, i, err
			}

			if i < len(r) && r[i] == comma {
				i++
			}

			if col+val >= len(d.arr) {
				d.arr = ensure(d.arr, col+val)

				if tag == cbor.Map && hdr == 0 {
					d.arr[col] = b.AppendValue(fmt.Sprintf("c%02d", col/(1+val)))
				}
			}

			d.arr[col+val-hdr] = off
			col += 1 + val
		}

		for i < len(r) && (r[i] == '\n' || r[i] == '\r') {
			i++
		}

		for col < len(d.arr) {
			d.arr[col+val] = jq.Null
		}

		if d.Header && !d.header {
			d.header = true
			continue
		}

		off = b.Writer().ArrayMap(tag, d.arr)

		return off, i, nil
	}
}

func (d *Decoder) decodeOne(b *jq.Buffer, r []byte, st int) (off jq.Off, i int, err error) {
	str := cbor.String
	comma := csel(d.Comma != 0, d.Comma, ',')
	i = st

	if r[i] == comma || r[i] == '\n' || r[i] == '\r' {
		return jq.Null, i, nil
	}

	q := r[i] == '"'

	bw := b.Writer()
	off = bw.Off()

	expl := 20
	b.B = b.Encoder.CBOR.AppendTag(b.B, str, expl)
	data := len(b.B)
	b.B, i = d.unquote(b.B, r, i)
	b.B = b.Encoder.CBOR.InsertLen(b.B, str, data, expl, len(b.B)-data)

	if val := d.decodeVal(b, b.B[data:], q); val != jq.None {
		if val < 0 {
			bw.Reset(off)
			return val, i, nil
		}

		copy(b.B[off:], b.B[val:])
		b.B = b.B[:off+(jq.Off(len(b.B))-val)]

		return off, i, nil
	}

	return off, i, nil
}

func (d *Decoder) decodeVal(b *jq.Buffer, r []byte, q bool) (res jq.Off) {
	switch string(r) {
	case "":
		return csel(q, jq.EmptyString, jq.Null)
	case "null":
		return jq.Null
	case "false":
		return jq.False
	case "true":
		return jq.True
	case "0":
		return jq.Zero
	case "1":
		return jq.One
	}

	bw := b.Writer()

	x, err := strconv.ParseUint(string(r), 10, 64)
	if err == nil {
		return bw.Uint64(x)
	}

	y, err := strconv.ParseInt(string(r), 10, 64)
	if err == nil {
		return bw.Int64(y)
	}

	z, err := strconv.ParseFloat(string(r), 64)
	if err == nil {
		return bw.Float(z)
	}

	return jq.None
}

func (d *Decoder) unquote(w, r []byte, st int) (_ []byte, i int) {
	comma := csel(d.Comma != 0, d.Comma, ',')
	i = st

	if r[i] != '"' {
		for i < len(r) && r[i] != comma && r[i] != '\r' && r[i] != '\n' {
			i++
		}

		return append(w, r[st:i]...), i
	}

	i++

	for {
		done := i

		for i < len(r) && r[i] != '"' {
			i++
		}
		if i == len(r) {
			return w, -1
		}

		w = append(w, r[done:i]...)
		i++

		if i < len(r) && r[i] == '"' {
			w = append(w, '"')
			i++

			continue
		}

		break
	}

	return w, i
}

func (e *Decoder) String() string { return "@csvd" }

func ensure[T any](s []T, i int) []T {
	if cap(s) == 0 {
		return make([]T, i+1)
	}
	if i < len(s) {
		return s
	}

	var zero T

	for cap(s) < i+1 {
		s = append(s[:cap(s)], zero)
	}

	return s[:i+1]
}
