package jq

import (
	"nikand.dev/go/cbor"
)

type (
	Iter struct {
		IgnoreTypeError bool

		arr []int
		j   int
	}
)

var _ FilterPath = (*Iter)(nil)

func NewIter() *Iter { return &Iter{} }

func (f *Iter) ApplyToGetPath(b *Buffer, base Path, at int, next bool) (res int, path Path, at1 int, more bool, err error) {
	off := base[at]

	res, more, err = f.ApplyTo(b, off, next)
	if err != nil {
		return off, base, at, false, err
	}

	index := f.j
	if b.Reader().Tag(off) == cbor.Map {
		index /= 2
	}

	path = base
	at++

	return res, path, at, more, nil
}

func (f *Iter) ApplyTo(b *Buffer, off int, next bool) (_ int, more bool, err error) {
	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Array && tag != cbor.Map {
		if f.IgnoreTypeError {
			return None, false, nil
		}

		return None, false, ErrType
	}

	val := 0
	if tag == cbor.Map {
		val = 1
	}

	if !next {
		f.arr = br.ArrayMap(off, f.arr[:0])
		f.j = 0
	} else {
		f.j += 1 + val
	}

	//	log.Printf("buf r (%x) % x  w (%x) % x", len(b.r), b.r, len(b.w), b.w)
	//	log.Printf("tag %x  off %x  j %d  val %d  of %02x", tag, off, f.j, val, f.arr)

	if f.j == len(f.arr) {
		return None, false, nil
	}

	more = (f.j + 1 + val) < len(f.arr)

	return f.arr[f.j+val], more, nil
}

func (f Iter) String() string { return ".[]" }
