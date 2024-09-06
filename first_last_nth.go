package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	First struct {
		Expr Filter
	}

	Last struct {
		Expr Filter
	}

	Nth struct {
		N    int
		Expr Filter
	}
)

func NewFirst(e Filter) First    { return First{Expr: e} }
func NewLast(e Filter) Last      { return Last{Expr: e} }
func NewNth(n int, e Filter) Nth { return Nth{N: n, Expr: e} }

func (f First) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()

	if f.Expr == nil {
		tag := br.Tag(off)
		if tag != cbor.Array {
			return off, false, ErrType
		}

		_, res = br.ArrayMapIndex(off, 0)
		if res == None {
			res = Null
		}

		return res, false, nil
	}

	return f.Expr.ApplyTo(b, None, next)
}

func (f Last) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()

	if f.Expr == nil {
		tag := br.Tag(off)
		if tag != cbor.Array {
			return off, false, ErrType
		}

		_, res = br.ArrayMapIndex(off, -1)
		if res == None {
			res = Null
		}

		return res, false, nil
	}

	last := Off(Null)

	for {
		res, next, err = f.Expr.ApplyTo(b, None, next)
		if err != nil {
			return off, false, err
		}

		if res != None {
			last = res
		}

		if !next {
			break
		}
	}

	return last, false, nil
}

func (f Nth) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()

	if f.Expr == nil {
		tag := br.Tag(off)
		if tag != cbor.Array {
			return off, false, ErrType
		}

		_, res = br.ArrayMapIndex(off, f.N)
		if res == None {
			res = Null
		}

		return res, false, nil
	}

	n := 0

	for {
		res, next, err = f.Expr.ApplyTo(b, None, next)
		if err != nil {
			return off, false, err
		}

		if res != None && n == f.N {
			return res, false, nil
		}
		if res != None {
			n++
		}

		if !next {
			break
		}
	}

	return None, false, nil
}

func (f First) String() string {
	if f.Expr == nil {
		return "first"
	}

	return fmt.Sprintf("first(%v)", f.Expr)
}

func (f Last) String() string {
	if f.Expr == nil {
		return "last"
	}

	return fmt.Sprintf("last(%v)", f.Expr)
}

func (f Nth) String() string {
	if f.Expr == nil {
		return fmt.Sprintf("nth(%d)", f.N)
	}

	return fmt.Sprintf("nth(%v; %v)", f.N, f.Expr)
}
