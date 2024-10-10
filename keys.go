package jq

import "nikand.dev/go/cbor"

type (
	Keys struct {
		arr []Off
	}
)

func NewKeys() *Keys { return &Keys{} }

func (f *Keys) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()
	tag := br.Tag(off)
	if tag != cbor.Map {
		return off, false, fe(f, off, NewTypeError(tag, cbor.Map))
	}

	f.arr = br.ArrayMap(off, f.arr[:0])

	for i := 0; i < len(f.arr)/2; i++ {
		f.arr[i] = f.arr[2*i]
	}

	f.arr = f.arr[:len(f.arr)/2]
	b.SortArray(f.arr)

	res = b.Writer().Array(f.arr)

	return res, false, nil
}
