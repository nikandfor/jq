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

func (f *Pipe) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if len(f.Filters) == 0 {
		return off, false, nil
	}

	bw := b.Writer()

	reset := bw.Len()
	defer bw.ResetIfErr(reset, &err)

	if !next {
		f.init(off)
	}

	back := func(fi int) int {
		if !next {
			return 0
		}

		for ; fi >= 0; fi-- {
			st := f.stack[fi]
			if st.next {
				break
			}

			f.stack[fi] = pipeState{off: None}
		}

		return fi
	}

	fi := len(f.Filters) - 1

back:
	for {
		fi = back(fi)
		if fi < 0 {
			return None, false, nil
		}

		next = true

		for ; fi < len(f.Filters); fi++ {
			st := f.stack[fi]

			f.stack[fi+1].off, f.stack[fi].next, err = f.Filters[fi].ApplyTo(b, st.off, st.next)
			if err != nil {
				return None, false, err
			}

			if f.stack[fi+1].off == None {
				continue back
			}
		}

		break
	}

	off = f.stack[len(f.Filters)].off
	more = back(len(f.Filters)-1) >= 0

	//	log.Printf("pipe %x %v  %v", off, more, f.stack)

	return off, more, nil
}

func (f *Pipe) init(off int) {
	for cap(f.stack) < len(f.Filters)+1 {
		f.stack = append(f.stack[:cap(f.stack)], pipeState{})
	}

	f.stack = f.stack[:len(f.Filters)+1]

	f.stack[0].off = off
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
