package jq

import (
	"fmt"
	"strings"
)

type (
	Pipe struct {
		Filters []Filter

		stack []pipeState
		path  Path
	}

	pipeState struct {
		off  Off
		path int
		next bool
	}
)

var _ FilterPath = (*Pipe)(nil)

func NewPipe(fs ...Filter) *Pipe {
	return &Pipe{Filters: fs}
}

func (f *Pipe) ApplyToGetPath(b *Buffer, off Off, base Path, next bool) (res Off, path Path, more bool, err error) {
	return f.applyTo(b, off, base, next, true)
}

func (f *Pipe) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, more, err = f.applyTo(b, off, nil, next, false)
	return
}

func (f *Pipe) applyTo(b *Buffer, off Off, base Path, next, addpath bool) (res Off, path Path, more bool, err error) {
	if len(f.Filters) == 0 {
		return off, base, false, nil
	}

	bw := b.Writer()
	path = base

	reset := bw.Off()
	defer bw.ResetIfErr(reset, &err)

	if !next {
		f.init(off, base)
	} else {
		path = append(path, f.path...)
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

		for ; fi < len(f.Filters); fi++ {
			st := f.stack[fi]
			ff := f.Filters[fi]

			fp := filterPath(ff)

			if addpath && fp != nil {
				off, path, f.stack[fi].next, err = fp.ApplyToGetPath(b, st.off, path[:st.path], st.next)
			} else {
				off, f.stack[fi].next, err = ff.ApplyTo(b, st.off, st.next)
			}
			//	log.Printf("pipe step %d  %v:%v -> %v:%v  (%v)", fi, path[:st.path], st.off, path, off, addpath && fp != nil)
			if err != nil {
				return None, path, false, err
			}
			if off == None {
				continue back
			}

			f.stack[fi+1] = pipeState{
				off:  off,
				path: len(path),
			}
		}

		break
	}

	last := len(f.Filters)
	res = f.stack[last].off
	more = f.back(last) >= 0

	f.path = append(f.path[:0], path...)

	//	log.Printf("pipe %x %v  %v", off, more, f.stack)

	return res, path, more, nil
}

func (f *Pipe) init(off Off, path Path) {
	f.stack = resize(f.stack, len(f.Filters)+1)

	f.stack[0] = pipeState{off: off, path: len(path)}
}

func (f *Pipe) back(fi int) int {
	for fi--; fi >= 0; fi-- {
		if f.stack[fi].next {
			break
		}
	}

	return fi
}

func (s pipeState) String() string { return fmt.Sprintf("{%x %x %v}", s.off, s.path, s.next) }

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

func filterPath(f Filter) FilterPath {
	fp, ok := f.(FilterPath)
	if !ok {
		return nil
	}

	return fp
}

func resize[T any](s []T, n int) []T {
	if cap(s) == 0 {
		return make([]T, n)
	}

	var zero T

	for cap(s) < n {
		s = append(s[:cap(s)], zero)
	}

	return s[:n]
}

func csel[T any](c bool, t, f T) T {
	if c {
		return t
	}

	return f
}
