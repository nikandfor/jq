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

func (f *Comma) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if !next {
		f.j = 0
		f.next = false
	}

	res = None

	for f.j < len(f.Filters) {
		res, f.next, err = f.Filters[f.j].ApplyTo(b, off, f.next)
		if err != nil {
			return off, false, err
		}

		if !f.next {
			f.j++
		}

		if res != None {
			break
		}
	}

	return res, f.j < len(f.Filters), nil
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
