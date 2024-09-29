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

	total := 0

	for _, el := range f.arr {
		tag := br.Tag(el)
		if tag != cbor.String && tag != cbor.Bytes {
			return off, false, NewTypeError(tag, cbor.String, cbor.Bytes)
		}

		total += len(br.Bytes(el))
	}

	if len(f.arr) != 0 {
		total += len(sep) * (len(f.arr) - 1)
	}

	res = bw.Off()

	b.B = b.Encoder.CBOR.AppendTag(b.B, tag, total)

	for i, el := range f.arr {
		if i != 0 && len(sep) != 0 {
			b.B = append(b.B, sep...)
		}

		s := br.Bytes(el)
		b.B = append(b.B, s...)
	}

	return res, more, nil
}

func (j Join) String() string {
	return fmt.Sprintf("join(%v)", csel[any](j.Separator != nil, j.Separator, "null"))
}
