//go:build ignore

package jq

import (
	"fmt"
	"log"
)

type (
	Assign struct {
		LHS, RHS Index

		Relative bool
	}
)

func NewAssign(lhs, rhs []any, rel bool) *Assign {
	f := &Assign{
		Relative: rel,
	}

	f.LHS.Path = index(lhs)
	f.RHS.Path = index(rhs)

	return f
}

func (f *Assign) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	for {
		field, more, err := f.LHS.ApplyTo(b, off, next)
		if err != nil {
			return off, more, err
		}
		if field == None {
			break
		}

		base := off

		if f.Relative {
			base = field
		}

		res, _, err = f.RHS.ApplyTo(b, base, false)
		if err != nil {
			return off, false, err
		}

		log.Printf("assign %x = %x\nlstack %v", field, res, f.LHS.stack)
		log.Printf("buf\n%s", DumpBuffer(b))

		br := b.Reader()
		bw := b.Writer()

		for fi := len(f.LHS.Path) - 1; fi >= 0; fi-- {
			st := f.LHS.stack[fi]
			tag := br.Tag(st.off)

			base := len(f.LHS.arr)
			f.LHS.arr = br.ArrayMap(st.off, f.LHS.arr)

			idx := st.i + int(st.val) - st.st

			f.LHS.arr[base+idx] = res

			res = bw.ArrayMap(tag, f.LHS.arr[base:])
			f.LHS.stack[fi].off = res
		}

		if !more {
			break
		}

		next = true
	}

	return f.LHS.stack[0].off, more, nil
}

func (f *Assign) String() string {
	op := "="
	if f.Relative {
		op = "|="
	}

	return fmt.Sprintf("%v %s %v", f.LHS.String(), op, f.RHS.String())
}
