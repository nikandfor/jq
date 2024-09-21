package jq

import (
	"nikand.dev/go/cbor"
)

type (
	Any struct {
		arr []Off
	}

	All struct {
		arr []Off
	}
)

func NewAny() *Any { return &Any{} }
func NewAll() *All { return &All{} }

func (f *Any) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	res, f.arr, err = anyAllApplyTo(b, off, 0, f.arr[:0])
	return res, false, err
}

func (f *All) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	res, f.arr, err = anyAllApplyTo(b, off, 1, f.arr[:0])
	return res, false, err
}

func anyAllApplyTo(b *Buffer, off Off, flip Off, arr0 []Off) (res Off, arr []Off, err error) {
	arr = arr0

	tag := b.Reader().Tag(off)
	if tag != cbor.Array && tag != cbor.Map {
		return off, arr, NewTypeError(tag, cbor.Array, cbor.Map)
	}

	val := 0
	if tag == cbor.Map {
		val = 1
	}

	arr = b.Reader().ArrayMap(off, arr)

	for j := 0; j < len(arr); j += 1 + val {
		if (flip == 0) == IsTrue(b, arr[j+val]) {
			return True ^ flip, arr, nil
		}
	}

	return False ^ flip, arr, nil
}

func (Any) String() string { return "any" }
func (All) String() string { return "all" }
