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

	f.arr, err = ApplyGetAll(f.Of, b, off, f.arr[:0])
	if err != nil {
		return off, false, err
	}

	off = b.Writer().Array(f.arr)

	return off, false, nil
}

func ApplyGetAll(f Filter, b *Buffer, off int, arr []int) (arr0 []int, err error) {
	bw := b.Writer()

	defer bw.ResetIfErr(bw.Len(), &err)
	defer func(reset int) {
		if err != nil {
			arr0 = arr0[:reset]
		}
	}(len(arr))

	var sub int
	next := false

	for {
		sub, next, err = f.ApplyTo(b, off, next)
		if err != nil {
			return arr, err
		}

		if sub != None {
			arr = append(arr, sub)
		}

		if !next {
			break
		}
	}

	return arr, nil
}

func (f Array) String() string { return fmt.Sprintf("[%v]", f.Of) }
