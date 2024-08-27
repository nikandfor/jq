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

func (f *Alternate) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	subf := f.Left
	if f.right {
		subf = f.Right
	}

	first := !next

	//	defer func(next bool) {
	//		log.Printf("alternate  %x %v => %x %v   from %v", off, next, res, more, loc.Caller(1))
	//	}(next)

	for {
		res, more, err = subf.ApplyTo(b, off, next)
		if err != nil {
			return off, false, err
		}

		//	log.Printf("alternate  %x %v -> %x %v   left %v  f %v", off, next, res, more, !f.right, subf)

		if !f.right {
			if more && b.Reader().IsSimple(res, Null, False) {
				next = true
				continue
			}

			if !b.Reader().IsSimple(res, None, Null, False) {
				return res, more, nil
			}
		} else {
			if res != None {
				return res, more, nil
			}
		}

		if !first || f.right {
			return None, false, nil
		}

		f.right = true
		subf = f.Right
		next = false
	}
}

func (f Alternate) String() string {
	return fmt.Sprintf("%v // %v", f.Left, f.Right)
}
