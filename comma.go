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

func (f *Comma) ApplyToGetPath(b *Buffer, base Path, at int, next bool) (res Off, path Path, at1 int, more bool, err error) {
	return f.applyTo(b, -1, base, at, next)
}

func (f *Comma) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, _, more, err = f.applyTo(b, off, nil, -1, next)
	return
}

func (f *Comma) applyTo(b *Buffer, off Off, base Path, at int, next bool) (res Off, path Path, at1 int, more bool, err error) {
	if !next {
		f.j = 0
		f.next = false
	}

	res = None
	path = base
	at1 = at

	for f.j < len(f.Filters) {
		ff := f.Filters[f.j]

		fp := filterPath(ff, at)

		if fp != nil {
			res, path, at1, f.next, err = fp.ApplyToGetPath(b, path, at, f.next)
		} else {
			res, f.next, err = ff.ApplyTo(b, off, f.next)
		}
		if err != nil {
			return off, path, at, false, err
		}

		if !f.next {
			f.j++
		}

		if res != None {
			break
		}
	}

	return res, path, at1, f.j < len(f.Filters), nil
}

func (f Comma) String() string {
	if len(f.Filters) == 0 {
		return ""
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
