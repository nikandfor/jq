package jq

import (
	"fmt"
)

type (
	Alternate struct {
		Left, Right Filter

		right bool
	}
)

func NewAlternate(l, r Filter) *Alternate { return &Alternate{Left: l, Right: r} }

func (f *Alternate) ApplyToGetPath(b *Buffer, off Off, base Path, next bool) (res Off, path Path, more bool, err error) {
	return f.applyTo(b, off, base, next, true)
}

func (f *Alternate) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, more, err = f.applyTo(b, off, nil, next, false)
	return
}

func (f *Alternate) applyTo(b *Buffer, off Off, base Path, next, addpath bool) (res Off, path Path, more bool, err error) {
	subf := f.Left
	if f.right {
		subf = f.Right
	}

	first := !next
	path = base

	//	defer func(next bool) {
	//		log.Printf("alternate  %x %v => %x %v   from %v", off, next, res, more, loc.Caller(1))
	//	}(next)

	for {
		if addpath {
			res, path, more, err = ApplyGetPath(subf, b, off, base, next)
		} else {
			res, more, err = subf.ApplyTo(b, off, next)
		}
		if err != nil {
			return off, path, false, err
		}

		//	log.Printf("alternate  %x %v -> %x %v   left %v  f %v", off, next, res, more, !f.right, subf)

		if !f.right {
			if more && b.Reader().IsSimple(res, Null, False) {
				next = true
				continue
			}

			if !b.Reader().IsSimple(res, None, Null, False) {
				return res, path, more, nil
			}
		} else {
			if res != None {
				return res, path, more, nil
			}
		}

		if !first || f.right {
			return None, path, false, nil
		}

		f.right = true
		subf = f.Right
		next = false
	}
}

func (f Alternate) String() string {
	return fmt.Sprintf("%v // %v", f.Left, f.Right)
}
