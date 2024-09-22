package jq

import (
	"fmt"
	"strings"
)

type (
	Comma struct {
		Filters []Filter

		j    int
		next bool
	}
)

var _ FilterPath = (*Comma)(nil)

func NewComma(fs ...Filter) *Comma {
	return &Comma{Filters: fs}
}

func (f *Comma) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	return f.applyTo(b, off, base, next, true)
}

func (f *Comma) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, more, err = f.applyTo(b, off, nil, next, false)
	return
}

func (f *Comma) applyTo(b *Buffer, off Off, base NodePath, next, addpath bool) (res Off, path NodePath, more bool, err error) {
	if !next {
		f.j = 0
		f.next = false
	}

	res = None
	path = base

	for f.j < len(f.Filters) {
		ff := f.Filters[f.j]

		if addpath {
			res, path, f.next, err = ApplyGetPath(ff, b, off, path, f.next)
		} else {
			res, f.next, err = ff.ApplyTo(b, off, f.next)
		}
		if err != nil {
			return off, path, false, err
		}

		if !f.next {
			f.j++
		}

		if res != None {
			break
		}
	}

	return res, path, f.j < len(f.Filters), nil
}

func (f Comma) String() string {
	if len(f.Filters) == 0 {
		return "empty"
	}

	var b strings.Builder

	_ = b.WriteByte('(')

	for i, sub := range f.Filters {
		if i != 0 {
			_, _ = b.WriteString(", ")
		}

		_, _ = fmt.Fprintf(&b, "%v", sub)
	}

	_ = b.WriteByte(')')

	return b.String()
}
