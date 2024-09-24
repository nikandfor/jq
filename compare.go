package jq

import (
	"fmt"
)

type (
	Equal struct {
		L, R Filter
		Not  bool

		binop Binop
	}

	Binop struct {
		lastl        Off
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

func (f *Equal) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	l, r, more, err := f.binop.ApplyTo(b, off, next, f.L, f.R)
	if err != nil || l == None {
		return None, more, err
	}

	if b.Equal(l, r) == !f.Not {
		return True, more, nil
	}

	return False, more, nil
}

func (f *Binop) ApplyTo(b *Buffer, off Off, next bool, L, R Filter) (l, r Off, more bool, err error) {
	if !next {
		f.lastl = None
		f.lnext = false
		f.rnext = false
	} else if !f.lnext && !f.rnext {
		return None, None, false, nil
	}

	//	defer func(off Off, next bool, ff Equal) {
	//		log.Printf("cmp equal %x %v (%x %v %v) -> %x %v  %x %v -> %v %v", off, next, ff.lastl, ff.lnext, ff.rnext, f.lastl, f.lnext, r, f.rnext, b.Equal(f.lastl, r) == !f.Not, more)
	//	}(off, next, *f)

	for !next || f.lnext || f.rnext {
		next = true

		if !f.rnext {
			f.lastl, f.lnext, err = L.ApplyTo(b, off, f.lnext)
			if err != nil {
				return off, off, false, err
			}

			if f.lastl == None {
				continue
			}
		}

		r, f.rnext, err = R.ApplyTo(b, off, f.rnext)
		if err != nil {
			return off, off, false, err
		}

		if r == None {
			continue
		}

		break
	}

	if f.lastl == None || r == None {
		return None, None, false, nil
	}

	more = f.lnext || f.rnext

	return f.lastl, r, more, nil
}

func NewNot() Not { return Not{} }

func (f Not) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next || off == None {
		return None, false, nil
	}

	if IsTrue(b, off) {
		return False, more, nil
	}

	return True, more, nil
}

func NewNotOf(f Filter) NotOf { return NotOf{Of: f} }

func (f NotOf) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
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
