package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Path struct {
		Expr FilterPath

		path NodePath
		arr  []Off
	}
)

func NewPath(e FilterPath) *Path { return &Path{Expr: e} }

func (f *Path) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, f.path, more, err = f.Expr.ApplyToGetPath(b, off, f.path[:0], next)
	if err != nil {
		return off, false, err
	}
	if res == None {
		return None, more, nil
	}

	br := b.Reader()
	bw := b.Writer()

	f.arr = f.arr[:0]

	for _, ps := range f.path {
		tag := br.Tag(ps.Off)

		var val Off

		switch tag {
		case cbor.Array:
			val = bw.Int(ps.Index)
		case cbor.Map:
			val, _ = br.ArrayMapIndex(ps.Off, ps.Index)
		default:
			panic(tag)
		}

		f.arr = append(f.arr, val)
	}

	res = bw.Array(f.arr)

	return res, more, nil
}

func (p Path) String() string { return fmt.Sprintf("path(%v)", p.Expr) }
