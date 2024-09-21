package jqcbor

import (
	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Off = jq.Off

	Decoder struct {
		CBOR cbor.Decoder

		arr []Off
	}
)

func NewDecoder() *Decoder {
	return &Decoder{
		CBOR: cbor.Decoder{Flags: cbor.FtDefault},
	}
}

func (d *Decoder) ApplyTo(b *jq.Buffer, off Off, next bool) (Off, bool, error) {
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

func (d *Decoder) Decode(b *jq.Buffer, r []byte, st int) (off Off, i int, err error) {
	bw := b.Writer()

	reset := bw.Off()
	defer bw.ResetIfErr(reset, &err)

	tag, sub, i := d.CBOR.Tag(r, st)

	switch tag {
	case cbor.Int, cbor.Neg, cbor.Bytes, cbor.String, cbor.Simple:
		switch r[st] {
		case cbor.Int | 0:
			return jq.Zero, i, nil
		case cbor.Int | 1:
			return jq.One, i, nil
		case cbor.Simple | cbor.Null:
			return jq.Null, i, nil
		case cbor.Simple | cbor.True:
			return jq.True, i, nil
		case cbor.Simple | cbor.False:
			return jq.False, i, nil
		}

		if tag == cbor.Bytes || tag == cbor.String {
			i += int(sub)
		}

		off = bw.Raw(r[st:i])

		return off, i, nil
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	arrbase := len(d.arr)
	defer func() { d.arr = d.arr[:arrbase] }()

	val := 0
	if tag == cbor.Map {
		val = 1
	}

	for j := 0; j < int(sub) || sub < 0 && !d.CBOR.Break(r, &i); j++ {
		for v := 0; v <= val; v++ {
			off, i, err = d.Decode(b, r, i)
			if err != nil {
				return
			}

			d.arr = append(d.arr, off)
		}
	}

	off = bw.ArrayMap(tag, d.arr[arrbase:])

	return off, i, nil
}
