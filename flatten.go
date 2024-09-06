package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Flatten struct {
		MaxDepth int

		arr []Off
		res []Off
	}
)

func NewFlatten(depth int) *Flatten { return &Flatten{MaxDepth: depth} }

func (f *Flatten) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	if f.MaxDepth == 0 {
		return off, false, nil
	}

	f.res = f.apply(b, off, 0, f.res[:0])

	res = b.Writer().Array(f.res)

	return res, false, nil
}

func (f *Flatten) apply(b *Buffer, off Off, depth int, res []Off) []Off {
	if f.MaxDepth >= 0 && depth > f.MaxDepth {
		return append(res, off)
	}

	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Array {
		return append(res, off)
	}

	arrbase := len(f.arr)
	defer func() { f.arr = f.arr[:arrbase] }()

	f.arr = br.ArrayMap(off, f.arr)

	for j := arrbase; j < len(f.arr); j++ {
		res = f.apply(b, f.arr[j], depth+1, res)
	}

	return res
}

func (f *Flatten) String() string {
	if f.MaxDepth < 0 {
		return "flatten"
	}

	return fmt.Sprintf("flatten(%d)", f.MaxDepth)
}
