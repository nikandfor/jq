package jqtext

import (
	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Codec struct {
		Tag jq.Tag

		Separator []byte

		sep bool
	}
)

func New() *Codec {
	return &Codec{
		Tag: cbor.Bytes,
	}
}

func (c *Codec) ApplyTo(b *jq.Buffer, off jq.Off, next bool) (jq.Off, bool, error) {
	if next {
		return jq.None, false, nil
	}

	tag := b.Reader().Tag(off)
	if tag != cbor.Bytes && tag != cbor.String {
		return off, false, jq.NewTypeError(tag, cbor.Bytes, cbor.String)
	}

	tag1 := csel(c.Tag == cbor.Bytes, cbor.Bytes, cbor.String)
	if tag1 == tag {
		return off, false, nil
	}

	s := b.Reader().Bytes(off)
	res := b.Writer().TagBytes(tag1, s)

	return res, false, nil
}

func (c *Codec) Decode(b *jq.Buffer, r []byte, st int) (off jq.Off, i int, err error) {
	tag := csel(c.Tag == cbor.Bytes, cbor.Bytes, cbor.String)
	off = b.Writer().TagBytes(tag, r[st:])

	return off, len(r), nil
}

func (c *Codec) Encode(w []byte, b *jq.Buffer, off jq.Off) (_ []byte, err error) {
	tag := b.Reader().Tag(off)
	if tag != cbor.Bytes && tag != cbor.String {
		return w, jq.NewTypeError(tag, cbor.Bytes, cbor.String)
	}

	if b.Equal(off, jq.Null) {
		return w, nil
	}

	s := b.Reader().Bytes(off)

	if c.sep {
		w = append(w, c.Separator...)
	}

	c.sep = true

	return append(w, s...), nil
}

func (c *Codec) Reset() {
	c.sep = false
}

func csel[T any](c bool, t, f T) T {
	if c {
		return t
	}

	return f
}
