package jq

import "fmt"

type (
	Array struct {
		Of Filter

		arr []Off
	}
)

func NewArray(of Filter) *Array { return &Array{Of: of} }

func (f *Array) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
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

func ApplyGetAll(f Filter, b *Buffer, off Off, arr []Off) (res []Off, err error) {
	bw := b.Writer()
	res = arr

	defer bw.ResetIfErr(bw.Off(), &err)
	defer func(reset int) {
		if err != nil {
			res = res[:reset]
		}
	}(len(res))

	var sub Off
	next := false

	for {
		sub, next, err = f.ApplyTo(b, off, next)
		if err != nil {
			return res, err
		}

		if sub != None {
			res = append(res, sub)
		}

		if !next {
			break
		}
	}

	return res, nil
}

func ApplyFuncAll(f Filter, b *Buffer, off Off, p func(off Off) error) (err error) {
	var sub Off
	next := false

	for {
		sub, next, err = f.ApplyTo(b, off, next)
		if err != nil {
			return err
		}

		if sub != None {
			err = p(sub)
			if err != nil {
				return err
			}
		}

		if !next {
			break
		}
	}

	return nil
}

func (f Array) String() string { return fmt.Sprintf("[%v]", f.Of) }
