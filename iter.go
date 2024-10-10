package jq

import (
	"nikand.dev/go/cbor"
)

type (
	Iter struct {
		NoError bool

		arr []Off
		j   int
	}
)

var _ FilterPath = (*Iter)(nil)

func NewIter() *Iter        { return &Iter{} }
func NewIterNoError() *Iter { return &Iter{NoError: true} }
func NewIterOf(f Filter) *Pipe {
	return NewPipe(f, NewIter())
}

func (f *Iter) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	return f.applyTo(b, off, base, next, true)
}

func (f *Iter) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, more, err = f.applyTo(b, off, nil, next, false)
	return
}

func (f *Iter) applyTo(b *Buffer, off Off, base NodePath, next, addpath bool) (res Off, path NodePath, more bool, err error) {
	br := b.Reader()
	tag := br.Tag(off)

	path = base
	val := csel(tag == cbor.Map, 1, 0)

	if !next {
		if tag != cbor.Array && tag != cbor.Map {
			if f.NoError {
				return None, base, false, nil
			}

			err = NewTypeError(tag, cbor.Array, cbor.Map)
			return None, base, false, fe(f, off, err)
		}

		f.arr = br.ArrayMap(off, f.arr[:0])
		f.j = 0
	} else {
		f.j += 1 + val
	}

	//	log.Printf("buf r (%x) % x  w (%x) % x", len(b.r), b.r, len(b.w), b.w)
	//	log.Printf("tag %x  off %x  j %d  val %d  of %02x", tag, off, f.j, val, f.arr)

	if f.j >= len(f.arr) {
		return None, path, false, nil
	}

	more = (f.j + 1 + val) < len(f.arr)

	if addpath {
		path = append(path, NodePathSeg{Off: off, Index: f.j / (1 + val), Key: None})
	}

	return f.arr[f.j+val], path, more, nil
}

func (f Iter) String() string { return ".[]" }
