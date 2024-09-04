//go:build ignore

package jq

import (
	"fmt"
	"log"
)

type (
	Assign struct {
		L FilterPath
		R Filter

		Relative bool

		field int
		path  Path

		state []indexState
		arr   []int
	}
)

func NewAssign(l FilterPath, r Filter, rel bool) *Assign {
	f := &Assign{
		L: l, R: r,
		Relative: rel,
	}

	return f
}

func (f *Assign) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	br := b.Reader()
	_ = br

	if !next {
		f.state = f.state[:0]
		f.path = f.path[:0]
	}

	f.field, f.path, more, err = f.L.ApplyToGetPath(b, off, next, f.path)
	if err != nil {
		return off, false, err
	}

	f.state = resize(f.state, len(f.path))

	pref := len(f.path)
	for pref--; pref >= 0; pref-- {
		st := f.state[pref]
		_ = st
	}

	rbase := csel(f.Relative, res, off)

	val, _, err := f.R.ApplyTo(b, rbase, false)
	if err != nil {
		return off, false, err
	}

	log.Printf("assign %x/%x = %x", f.path, f.field, val)

	return f.path[0].Off, false, nil
}

func (f *Assign) String() string {
	op := "="
	if f.Relative {
		op = "|="
	}

	return fmt.Sprintf("%v %s %v", f.L, op, f.R)
}

func resize[T any](s []T, n int) []T {
	if cap(s) < n {
		return make([]T, n)
	}

	return s[:n]
}
