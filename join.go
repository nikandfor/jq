package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Join struct {
		Separator Filter
		Tag       Tag

		arr []Off
	}
)

func NewJoin(sep Filter) *Join { return &Join{Separator: sep, Tag: cbor.String} }
func NewJoinOf(f, sep Filter) *Pipe {
	return NewPipe(f, NewJoin(sep))
}

func (f *Join) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	br := b.Reader()
	bw := b.Writer()
	sep := []byte{}

	if tag := br.Tag(off); tag != cbor.Array {
		return off, false, NewTypeError(tag, cbor.Array)
	}

	if f.Separator != nil {
		res, more, err = f.Separator.ApplyTo(b, off, next)
		if err != nil {
			return off, false, err
		}
		if res == None {
			return None, more, nil
		}

		sep = br.Bytes(res)
	}

	tag := csel(f.Tag != 0, f.Tag, cbor.String)
	f.arr = br.ArrayMap(off, f.arr[:0])

	if len(f.arr) == 0 && tag == cbor.String {
		return EmptyString, more, nil
	}

	res = bw.Off()
	expl := 100

	b.B = b.Encoder.CBOR.AppendTag(b.B, tag, expl)
	st := len(b.B)

	for i, el := range f.arr {
		if i != 0 && len(sep) != 0 {
			b.B = append(b.B, sep...)
		}

		tag := br.Tag(el)
		if tag != cbor.String && tag != cbor.Bytes {
			return off, false, NewTypeError(tag, cbor.String, cbor.Bytes)
		}

		s := br.Bytes(el)
		b.B = append(b.B, s...)
	}

	b.B = b.Encoder.CBOR.InsertLen(b.B, tag, st, expl, len(b.B)-st)

	return res, more, nil
}

func (j Join) String() string {
	return fmt.Sprintf("join(%v)", csel[any](j.Separator != nil, j.Separator, "null"))
}
