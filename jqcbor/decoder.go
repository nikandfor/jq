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
		return off, false, jq.NewTypeError(tag, cbor.Bytes, cbor.String)
	}

	s := br.Bytes(off)

	res, _, err := d.Decode(b, s, 0)
	if err != nil {
		return off, false, err
	}

	return res, false, nil
}

func (d *Decoder) Decode(b *jq.Buffer, r []byte, st int) (off Off, i int, err error) {
	switch cbor.Tag(r[st]) {
	case cbor.Int | 0:
		return jq.Zero, st + 1, nil
	case cbor.Int | 1:
		return jq.One, st + 1, nil
	case cbor.Simple | cbor.Null:
		return jq.Null, st + 1, nil
	case cbor.Simple | cbor.True:
		return jq.True, st + 1, nil
	case cbor.Simple | cbor.False:
		return jq.False, st + 1, nil
	}

	bw := b.Writer()

	reset := bw.Off()
	defer bw.ResetIfErr(reset, &err)

	i = st

	off, i = d.decodeLabels(b, r, i)
	labels := i != st

	tag, sub, end := d.CBOR.Tag(r, i)

	switch tag {
	case cbor.Int, cbor.Neg, cbor.Bytes, cbor.String, cbor.Simple:
		if tag == cbor.Bytes || tag == cbor.String {
			end += int(sub)
		}

		_ = bw.Raw(r[i:end])

		return reset, end, nil
	case cbor.Array, cbor.Map:
		i = end
		bw.Reset(reset)
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

	if len(d.arr[arrbase:]) == 0 && !labels {
		return jq.EmptyArray, i, nil
	}

	off, _ = d.decodeLabels(b, r, st)
	_ = bw.ArrayMap(tag, d.arr[arrbase:])

	return off, i, nil
}

func (d *Decoder) decodeLabels(b *jq.Buffer, r []byte, i int) (Off, int) {
	bw := b.Writer()
	off := bw.Off()

	for {
		tag, sub, end := d.CBOR.Tag(r, i)
		if tag != cbor.Label {
			break
		}

		bw.Label(jq.LabeledOffset + int(sub))
		i = end
	}

	return off, i
}

func (d *Decoder) String() string { return "@cbord" }
