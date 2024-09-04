package jq

import (
	"fmt"
	"strings"
)

type (
	Pipe struct {
		Filters []Filter

		stack []pipeState
	}

	pipeState struct {
		off  int
		next bool
	}
)

func NewPipe(fs ...Filter) *Pipe {
	return &Pipe{Filters: fs}
}

func (f *Pipe) ApplyToGetPath(b *Buffer, off int, next bool, base Path) (res int, path Path, more bool, err error) {
	return f.applyToGetPath(b, off, next, base, true)
}

func (f *Pipe) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	res, _, more, err = f.applyToGetPath(b, off, next, nil, false)

	return res, more, err
}

func (f *Pipe) applyToGetPath(b *Buffer, off int, next bool, base Path, usepath bool) (res int, path Path, more bool, err error) {
	if len(f.Filters) == 0 {
		return off, base, false, nil
	}

	bw := b.Writer()

	reset := bw.Len()
	defer bw.ResetIfErr(reset, &err)

	path = base
	pathReset := len(path)
	defer func() {
		if err != nil {
			path = path[:pathReset]
		}
	}()

	if !next {
		f.init(off)
	}

	fi := len(f.Filters)

back:
	for {
		fi = f.back(fi)
		if !next {
			fi = 0
			next = true
		}
		if fi < 0 {
			return None, base, false, nil
		}

		path = path[:pathReset]

		for ; fi < len(f.Filters); fi++ {
			st := f.stack[fi]
			ff := f.Filters[fi]

			fp := filterPath(ff, usepath)

			if fp != nil {
				f.stack[fi+1].off, path, f.stack[fi].next, err = fp.ApplyToGetPath(b, st.off, st.next, path)
			} else {
				f.stack[fi+1].off, f.stack[fi].next, err = ff.ApplyTo(b, st.off, st.next)
			}
			if err != nil {
				return None, base, false, err
			}
			if f.stack[fi+1].off == None {
				continue back
			}
		}

		break
	}

	res = f.stack[len(f.Filters)].off
	more = f.back(len(f.Filters)) >= 0

	//	log.Printf("pipe %x %v  %v", off, more, f.stack)

	return res, path, more, nil
}

func (f *Pipe) init(off int) {
	f.stack = resize(f.stack, len(f.Filters)+1)
	f.stack[0].off = off
}

func (f *Pipe) back(fi int) int {
	for fi--; fi >= 0; fi-- {
		if f.stack[fi].next {
			break
		}
	}

	return fi
}

func (s pipeState) String() string { return fmt.Sprintf("{%x %v}", s.off, s.next) }

func (f Pipe) String() string {
	if len(f.Filters) == 0 {
		return "Pipe()"
	}

	var b strings.Builder

	_ = b.WriteByte('(')

	for i, sub := range f.Filters {
		if i != 0 {
			_, _ = b.WriteString(" | ")
		}

		_, _ = fmt.Fprintf(&b, "%v", sub)
	}

	_ = b.WriteByte(')')

	return b.String()
}

func filterPath(f Filter, usepath bool) FilterPath {
	if !usepath {
		return nil
	}

	fp, ok := f.(FilterPath)
	if !ok {
		return nil
	}

	return fp
}

func resize[T any](s []T, n int) []T {
	if cap(s) < n {
		return make([]T, n)
	}

	return s[:n]
}

func csel[T any](c bool, t, f T) T {
	if c {
		return t
	}

	return f
}
