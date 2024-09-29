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

		s, d.b, _, i = skip.DecodeString(r, i, skip.URL, d.b)
		if s.Err() {
			return jq.None, i, s
		}

		key := bw.TagBytes(cbor.String, d.b[br:])

		for j := 0; j < len(d.arr); j += 2 {
			if !b.Equal(d.arr[j], key) {
				continue
			}

			bw.Reset(key)
			key = d.arr[j]
			break
		}

		val := jq.Null
		br = len(d.b)

		if i < len(r) && r[i] == '=' {
			i++

			s, d.b, _, i = skip.DecodeString(r, i, skip.URL, d.b)
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

	dups := func(arr []jq.Off, j int) []jq.Off {
		for k := j; k < len(arr); k += 2 {
			if arr[k] == arr[j] {
				arr = append(arr, arr[k+1])
			}
		}

		return arr
	}

	move := func(arr []jq.Off, j int) []jq.Off {
		w := j + 2
		r := j + 2

		for ; r < len(arr); r += 2 {
			if arr[r] == arr[j] {
				continue
			}

			arr[w], arr[w+1] = arr[r], arr[r+1]
			w += 2
		}

		return arr[:w]
	}

	for j := 0; j < len(d.arr)-2; j += 2 {
		l := len(d.arr)

		d.arr = dups(d.arr, j)

		if len(d.arr) == l+1 {
			d.arr = d.arr[:l]
			continue
		}

		d.arr[j+1] = bw.Array(d.arr[l:])

		d.arr = move(d.arr[:l], j)
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
