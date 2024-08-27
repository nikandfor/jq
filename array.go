package jq

import "fmt"

type (
	Array struct {
		Of Filter

		arr []int
	}
)

func NewArray(of Filter) *Array { return &Array{Of: of} }

func (f *Array) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if next {
		return None, false, nil
	}

	bw := b.Writer()

	reset := bw.Len()
	defer bw.ResetIfErr(reset, &err)

	f.arr = f.arr[:0]
	next = false

	for {
		sub, more, err := f.Of.ApplyTo(b, off, next)
		next = more
		if err != nil {
			return off, false, err
		}

		if sub != None {
			f.arr = append(f.arr, sub)
		}

		if !more {
			break
		}
	}

	off = bw.Array(f.arr)

	return off, false, nil
}

func (f Array) String() string { return fmt.Sprintf("[%v]", f.Of) }
