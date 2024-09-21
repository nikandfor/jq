package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Slice struct {
		Low, High int

		arr []Off
	}
)

func NewSlice(lo, hi int) *Slice { return &Slice{Low: lo, High: hi} }

func (f *Slice) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	tag := b.Reader().Tag(off)
	if tag != cbor.Array {
		return off, false, NewTypeError(tag, cbor.Array)
	}

	f.arr = b.Reader().ArrayMap(off, f.arr[:0])
	l := len(f.arr)

	lo, hi := f.Low, f.High

	if lo < 0 {
		lo = l + lo
	}
	if hi < 0 {
		hi = l + hi
	}

	lo = bound(lo, 0, l)
	hi = bound(hi, 0, l)

	if lo == 0 && hi == l {
		return off, false, nil
	}

	if hi < lo {
		f.arr = append(f.arr, f.arr[:hi]...)
		hi += l
	}

	res = b.Writer().ArrayMap(tag, f.arr[lo:hi])

	return res, false, nil
}

func (f Slice) String() string { return fmt.Sprintf(".[%d:%d]", f.Low, f.High) }

func bound(x, lo, hi int) int {
	if x < lo {
		x = lo
	}
	if x > hi {
		x = hi
	}
	return x
}
