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

func NewComma(fs ...Filter) *Comma {
	return &Comma{Filters: fs}
}

func (f *Comma) ApplyToGetPath(b *Buffer, off int, next bool, base Path) (res int, path Path, more bool, err error) {
	return f.applyTo(b, off, next, base, true)
}

func (f *Comma) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	res, _, more, err = f.applyTo(b, off, next, nil, false)
	return
}

func (f *Comma) applyTo(b *Buffer, off int, next bool, base Path, usepath bool) (res int, path Path, more bool, err error) {
	if !next {
		f.j = 0
		f.next = false
	}

	path = base
	pathReset := len(path)
	defer func() {
		if err != nil {
			path = path[:pathReset]
		}
	}()

	res = None

	for f.j < len(f.Filters) {
		ff := f.Filters[f.j]

		fp := filterPath(ff, usepath)

		if fp != nil {
			res, path, f.next, err = fp.ApplyToGetPath(b, off, f.next, path)
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
