package jq

import (
	"fmt"
)

type (
	Equal struct {
		L, R Filter
		Not  bool

		lastl        int
		lnext, rnext bool
	}

	Not   struct{}
	NotOf struct {
		Of Filter
	}
)

func NewEqual(l, r Filter) *Equal {
	return &Equal{L: l, R: r}
}

func NewNotEqual(l, r Filter) *Equal {
	return &Equal{L: l, R: r, Not: true}
}

func (f *Equal) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if !next {
		f.lastl = None
		f.lnext = false
		f.rnext = false
	} else if !f.lnext && !f.rnext {
		return None, false, nil
	}

	var r int

	//	defer func(off int, next bool, ff Equal) {
	//		log.Printf("cmp equal %x %v (%x %v %v) -> %x %v  %x %v -> %v %v", off, next, ff.lastl, ff.lnext, ff.rnext, f.lastl, f.lnext, r, f.rnext, b.Equal(f.lastl, r) == !f.Not, more)
	//	}(off, next, *f)

	for !next || f.lnext || f.rnext {
		next = true

		if !f.rnext {
			f.lastl, f.lnext, err = f.L.ApplyTo(b, off, f.lnext)
			if err != nil {
				return off, false, err
			}

			if f.lastl == None {
				continue
			}
		}

		r, f.rnext, err = f.R.ApplyTo(b, off, f.rnext)
		if err != nil {
			return off, false, err
		}

		if r == None {
			continue
		}

		break
	}

	if f.lastl == None || r == None {
		return None, false, nil
	}

	more = f.lnext || f.rnext

	if b.Equal(f.lastl, r) == !f.Not {
		return True, more, nil
	}

	return False, more, nil
}

func NewNot() Not { return Not{} }

func (f Not) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if next || off == None {
		return None, false, nil
	}

	if IsTrue(b, off) {
		return False, more, nil
	}

	return True, more, nil
}

func NewNotOf(f Filter) NotOf { return NotOf{Of: f} }

func (f NotOf) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if f.Of == nil {
		return Not{}.ApplyTo(b, off, next)
	}

	off, more, err = f.Of.ApplyTo(b, off, next)
	if err != nil {
		return off, false, err
	}
	if off == None {
		return None, false, nil
	}

	if IsTrue(b, off) {
		return False, more, nil
	}

	return True, more, nil
}

func (f Equal) String() string {
	eq := "=="
	if f.Not {
		eq = "!="
	}

	return fmt.Sprintf("%v %s %v", f.L, eq, f.R)
}

func (f Not) String() string   { return "not" }
func (f NotOf) String() string { return fmt.Sprintf("not_of(%v)", f.Of) }
