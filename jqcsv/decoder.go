package jqcsv

import (
	"fmt"
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
	"nikand.dev/go/skip"
)

type (
	Decoder struct {
		Tag    jq.Tag
		Header bool
		Flags  skip.Str

		header bool

		arr []jq.Off
	}
)

func NewDecoder() *Decoder {
	return &Decoder{
		Tag:   cbor.Array,
		Flags: skip.CSV | skip.Raw | skip.Quo | skip.Sqt | ',',
	}
}

func (d *Decoder) Reset() {
	if d.Flags == 0 {
		d.Flags = skip.CSV | skip.Raw | skip.Quo | skip.Sqt | ','
	}

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
	tag := csel(d.Tag != 0, d.Tag, cbor.Array)
	val := csel(tag == cbor.Map, 1, 0)
	i = st

	if i == len(r) {
		return jq.None, i, nil
	}

	for {
		col := 0
		hdr := csel(d.Header && !d.header && tag == cbor.Map, 1, 0)

		for i < len(r) {
			off, i, err = d.decodeCell(b, r, i)
			if err != nil {
				return off, i, err
			}

			if col+val >= len(d.arr) {
				d.arr = ensure(d.arr, col+val)

				if tag == cbor.Map && hdr == 0 {
					d.arr[col] = b.AppendValue(fmt.Sprintf("c%02d", col/(1+val)))
				}
			}

			d.arr[col+val-hdr] = off
			col += 1 + val

			if i < len(r) && r[i] == '\r' {
				i++
			}
			if i < len(r) && r[i] == '\n' {
				i++
				break
			}
		}

		for col < len(d.arr) {
			d.arr[col+val] = jq.Null
			col += 1 + val
		}

		if d.Header && !d.header {
			d.header = true
			continue
		}

		off = b.Writer().ArrayMap(tag, d.arr)

		return off, i, nil
	}
}

func (d *Decoder) decodeCell(b *jq.Buffer, r []byte, st int) (off jq.Off, i int, err error) {
	str := cbor.String

	bw := b.Writer()
	off = bw.Off()

	expl := 20
	b.B = b.Encoder.CBOR.AppendTag(b.B, str, expl)
	data := len(b.B)

	var s skip.Str

	s, b.B, _, i = skip.DecodeString(r, st, d.Flags, b.B)
	if s.Err() {
		return jq.None, i, s
	}

	b.B = b.Encoder.CBOR.InsertLen(b.B, str, data, expl, len(b.B)-data)

	if val := d.decodeVal(b, b.B[data:], s.Any(skip.Quo|skip.Sqt)); val != jq.None {
		if val < 0 {
			bw.Reset(off)
			return val, i, nil
		}

		size := bw.Off() - val
		copy(b.B[off:], b.B[val:])
		b.B = b.B[:off+size]
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
