package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Nth int

	NthOf struct {
		Expr Filter
		N    int

		arr []Off
	}
)

func NewFirst() Nth    { return Nth(0) }
func NewLast() Nth     { return Nth(-1) }
func NewNth(n int) Nth { return Nth(n) }

func NewFirstOf(e Filter) NthOf      { return NthOf{Expr: e, N: 0} }
func NewLastOf(e Filter) NthOf       { return NthOf{Expr: e, N: -1} }
func NewNthOf(e Filter, n int) NthOf { return NthOf{Expr: e, N: n} }

func (f Nth) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Array && tag != cbor.Map {
		return off, false, ErrType
	}

	_, res = br.ArrayMapIndex(off, int(f))
	if res == None {
		res = Null
	}

	return res, false, nil
}

func (f NthOf) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	if n := f.N; n >= 0 {
		var next bool

		for {
			res, next, err = f.Expr.ApplyTo(b, off, next)
			if err != nil {
				return off, false, err
			}
			if res == None {
				continue
			}

			if n == 0 {
				return res, false, nil
			}

			if !next {
				break
			}

			n--
		}

		return None, false, nil
	}

	f.arr = resize(f.arr, -f.N)

	next = false
	i := 0

	for {
		res, next, err = f.Expr.ApplyTo(b, off, next)
		if err != nil {
			return off, false, err
		}
		if res == None {
			continue
		}

		f.arr[i%len(f.arr)] = res
		i++

		if !next {
			break
		}
	}

	i = i + f.N

	if i < 0 {
		return None, false, nil
	}

	res = f.arr[i%len(f.arr)]

	return res, false, nil
}

func (f Nth) String() string {
	if f == 0 {
		return "first"
	}

	if f == -1 {
		return "last"
	}

	return fmt.Sprintf("nth(%d)", int(f))
}

func (f NthOf) String() string {
	if f.N == 0 {
		return fmt.Sprintf("first(%v)", f.Expr)
	}

	if f.N == -1 {
		return fmt.Sprintf("last(%v)", f.Expr)
	}

	return fmt.Sprintf("nth(%d; %v)", f.N, f.Expr)
}
