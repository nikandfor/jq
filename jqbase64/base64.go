package jqbase64

import (
	"encoding/base64"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Decoder struct {
		Base64 *base64.Encoding

		Tag jq.Tag
	}

	Encoder struct {
		Base64 *base64.Encoding

		Tag       jq.Tag
		Separator []byte

		sep bool
	}
)

func NewDecoder(enc *base64.Encoding) *Decoder {
	return &Decoder{
		Base64: enc,
		Tag:    cbor.Bytes,
	}
}

func NewEncoder(enc *base64.Encoding) *Encoder {
	return &Encoder{
		Base64: enc,
		Tag:    cbor.String,
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
	enc := d.Base64
	if enc == nil {
		enc = base64.StdEncoding
	}

	tag := d.Tag
	if tag == 0 {
		tag = cbor.Bytes
	}

	n := enc.DecodedLen(len(r[st:]))

	off = b.Writer().Off()
	b.B = b.Encoder.CBOR.AppendTag(b.B, tag, n)

	mark := len(b.B)

	b.B, err = enc.AppendDecode(b.B, r[st:])
	if err != nil {
		return jq.None, st, err
	}

	b.B = b.Encoder.CBOR.InsertLen(b.B, tag, mark, n, len(b.B)-mark)

	return off, len(r), nil
}

func (e *Encoder) ApplyTo(b *jq.Buffer, off jq.Off, next bool) (jq.Off, bool, error) {
	if next {
		return jq.None, false, nil
	}

	var err error

	res := b.Writer().Off()

	tag := e.Tag
	if tag == 0 {
		tag = cbor.String
	}

	var ce cbor.Encoder

	expl := 100
	b.B = ce.AppendTag(b.B, tag, expl)
	st := len(b.B)

	b.B, err = e.Encode(b.B, b, off)
	if err != nil {
		return off, false, err
	}

	b.B = ce.InsertLen(b.B, tag, st, expl, len(b.B)-st)

	return res, false, nil
}

func (e *Encoder) Encode(w []byte, b *jq.Buffer, off jq.Off) (_ []byte, err error) {
	defer func(reset int) {
		if err != nil {
			w = w[:reset]
		}
	}(len(w))

	if e.sep {
		w = append(w, e.Separator...)
	}

	e.sep = true

	if b.Equal(off, jq.Null) {
		return w, nil
	}

	tag := b.Reader().Tag(off)
	if tag != cbor.Bytes && tag != cbor.String {
		return w, jq.NewTypeError(tag, cbor.Bytes, cbor.String)
	}

	s := b.Reader().Bytes(off)

	enc := e.Base64
	if enc == nil {
		enc = base64.StdEncoding
	}

	w = enc.AppendEncode(w, s)

	return w, nil
}

func (e *Encoder) Reset() {
	e.sep = false
}
