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
		off  Off
		at   int
		next bool
	}
)

var _ FilterPath = (*Pipe)(nil)

func NewPipe(fs ...Filter) *Pipe {
	return &Pipe{Filters: fs}
}

func (f *Pipe) ApplyToGetPath(b *Buffer, base Path, at int, next bool) (res Off, path Path, at1 int, more bool, err error) {
	return f.applyToGetPath(b, -1, base, at, next)
}

func (f *Pipe) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, _, more, err = f.applyToGetPath(b, off, nil, -1, next)

	return res, more, err
}

func (f *Pipe) applyToGetPath(b *Buffer, off Off, base Path, at int, next bool) (res Off, path Path, at1 int, more bool, err error) {
	if len(f.Filters) == 0 {
		return off, base, at, false, nil
	}

	bw := b.Writer()

	reset := bw.Len()
	defer bw.ResetIfErr(reset, &err)

	path = base

	if !next {
		f.init(off, path, at)
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
			return None, base, at, false, nil
		}

		for ; fi < len(f.Filters); fi++ {
			st := f.stack[fi]
			ff := f.Filters[fi]

			fp := filterPath(ff, at)

			//	log.Printf("pipe step %x %v  % x  %+v", fi, ff, path, st)

			if fp != nil {

				off, path, at, f.stack[fi].next, err = fp.ApplyToGetPath(b, path, st.at, st.next)
				f.stack[fi+1] = pipeState{
					off: off,
					at:  at,
				}
				path = append(path[:at], off)
			} else {
				f.stack[fi+1].off, f.stack[fi].next, err = ff.ApplyTo(b, st.off, st.next)
				f.stack[fi+1].at = st.at
			}
			//	log.Printf("pipe step %x %v   % x  %+v  %v  <----", fi, ff, path, f.stack[fi+1], err)
			if err != nil {
				return None, base, at, false, err
			}
			if f.stack[fi+1].off == None {
				continue back
			}
		}

		break
	}

	last := len(f.Filters)
	res = f.stack[last].off
	at = f.stack[last].at
	more = f.back(last) >= 0

	//	log.Printf("pipe %x %v  %v", off, more, f.stack)

	return res, path, at, more, nil
}

func (f *Pipe) init(off Off, path Path, at int) {
	f.stack = resize(f.stack, len(f.Filters)+1)

	if at < 0 {
		f.stack[0] = pipeState{off: off}
	} else {
		f.stack[0] = pipeState{off: path[at], at: at}
	}
}

func (f *Pipe) back(fi int) int {
	for fi--; fi >= 0; fi-- {
		if f.stack[fi].next {
			break
		}
	}

	return fi
}

func (s pipeState) String() string { return fmt.Sprintf("{%x %x %v}", s.off, s.at, s.next) }

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

func filterPath(f Filter, at int) FilterPath {
	if at < 0 {
		return nil
	}

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
