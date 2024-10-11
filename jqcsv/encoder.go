package jqcsv

import (
	"fmt"
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

type (
	Encoder struct {
		Tag     jq.Tag
		Comma   byte
		Newline []byte
		Null    []byte
		True    []byte
		False   []byte

		Marshaler jq.Filter

		MapHeader   bool
		ArrayHeader bool

		header bool
		arr    []jq.Off
	}
)

func NewEncoder() *Encoder {
	return &Encoder{
		Tag:   cbor.String,
		Comma: ',',
	}
}

func (e *Encoder) Reset() { e.header = false }

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

	br := b.Reader()
	tag := br.Tag(off)

	if !e.header && (tag == cbor.Map && e.MapHeader || tag == cbor.Array && e.ArrayHeader) {
		w, err = e.encodeLine(w, b, off, true)
		if err != nil {
			return w, err
		}

		e.header = true
	}

	return e.encodeLine(w, b, off, false)
}

func (e *Encoder) encodeLine(w []byte, b *jq.Buffer, off jq.Off, header bool) (_ []byte, err error) {
	br := b.Reader()
	tag := br.Tag(off)
	val := csel(tag == cbor.Map, 1, 0)
	hdr := csel(header && tag == cbor.Map, -1, 0)
	comma := csel(e.Comma != 0, e.Comma, ',')
	nl := csel(len(e.Newline) != 0, e.Newline, []byte{'\n'})

	e.arr = br.ArrayMap(off, e.arr[:0])

	for i := 0; i < len(e.arr); i += 1 + val {
		if i != 0 {
			w = append(w, comma)
		}

		w, err = e.encodeCell(w, b, e.arr[i+val+hdr])
		if err != nil {
			return w, err
		}
	}

	w = append(w, nl...)

	return w, nil
}

func (e *Encoder) encodeCell(w []byte, b *jq.Buffer, off jq.Off) (_ []byte, err error) {
	br := b.Reader()
	raw := br.TagRaw(off)
	tag := raw & cbor.TagMask

out:
	switch tag {
	case cbor.Int, cbor.Neg:
		v := br.Unsigned(off)

		if tag == cbor.Neg {
			w = append(w, '-')
		}

		return strconv.AppendUint(w, v, 10), nil
	case cbor.Bytes, cbor.String:
		s := br.Bytes(off)

		if tag == cbor.Bytes || e.needsQuote(s) {
			w = e.appendQuote(w, s)
		} else {
			w = append(w, s...)
		}

		return w, nil
	case cbor.Simple:
		var val []byte

		switch {
		case br.IsSimple(off, jq.Null):
			val = csel(e.Null != nil, e.Null, []byte(""))
		case br.IsSimple(off, jq.False):
			val = csel(e.False != nil, e.False, []byte("false"))
		case br.IsSimple(off, jq.True):
			val = csel(e.True != nil, e.True, []byte("true"))
		case cbor.IsFloat(raw):
			v := br.Float(off)
			w = strconv.AppendFloat(w, v, 'v', -1, 64)
			return w, nil
		default:
			break out
		}

		return append(w, val...), nil
	}

	if e.Marshaler == nil {
		err = jq.NewTypeError(raw, cbor.Simple|cbor.Float64, cbor.Simple|cbor.Null, cbor.Simple|cbor.True, cbor.Simple|cbor.False, cbor.Int, cbor.Neg, cbor.Bytes, cbor.String)
		return w, jq.NewError(e, off, err)
	}

	reset := b.Writer().Off()
	defer b.Writer().Reset(reset)

	res, _, err := e.Marshaler.ApplyTo(b, off, false)
	if err != nil {
		return w, fmt.Errorf("marshal: %w", err)
	}

	tag = br.Tag(res)
	if tag != cbor.Bytes && tag != cbor.String {
		err = jq.NewTypeError(tag, cbor.Bytes, cbor.String)
		return w, fmt.Errorf("marshal result: %w", err)
	}

	return e.encodeCell(w, b, res)
}

func (e *Encoder) appendQuote(w, s []byte) []byte {
	w = append(w, '"')

out:
	for i := 0; i < len(s); {
		for j, r := range string(s[i:]) {
			if r != '"' {
				continue
			}

			w = append(w, s[i:i+j]...)
			w = append(w, '"', '"')
			i += j + 1

			continue out
		}

		w = append(w, s[i:]...)
		break
	}

	return append(w, '"')
}

func (e *Encoder) needsQuote(s []byte) bool {
	comma := rune(csel(e.Comma != 0, e.Comma, ','))

	for _, r := range string(s) {
		if r == comma || r == '"' || r == '\n' || r == '\r' {
			return true
		}
	}

	return false
}

func (e *Encoder) String() string { return "@csv" }

func csel[T any](c bool, x, y T) T {
	if c {
		return x
	}

	return y
}
