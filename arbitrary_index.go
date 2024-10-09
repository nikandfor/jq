package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	ArbitraryIndex struct {
		Base  Filter
		Index Filter
		NoErr bool

		lastl        Off
		lnext, rnext bool

		lastp NodePath
	}
)

func NewIndex(base, index Filter) *ArbitraryIndex {
	return &ArbitraryIndex{Base: base, Index: index}
}

func NewIndexNoErr(base, index Filter) *ArbitraryIndex {
	return &ArbitraryIndex{Base: base, Index: index, NoErr: true}
}

func (f *ArbitraryIndex) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	return f.applyTo(b, off, base, next, true, f.NoErr)
}

func (f *ArbitraryIndex) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, more, err = f.applyTo(b, off, nil, next, false, f.NoErr)
	return
}

func (f *ArbitraryIndex) applyTo(b *Buffer, off Off, base NodePath, next bool, addpath addpath, null bool) (res Off, path NodePath, more bool, err error) {
	if !next {
		f.lastl = None
		f.lnext = false
		f.rnext = false
		f.lastp = f.lastp[:0]
	} else if !f.lnext && !f.rnext {
		return None, base, false, nil
	}

	path = base

	for !next || f.lnext || f.rnext {
		next = true

		if !f.rnext {
			if addpath {
				f.lastl, path, f.lnext, err = ApplyGetPath(f.Base, b, off, path[:len(base)], f.lnext)
			} else {
				f.lastl, f.lnext, err = f.Base.ApplyTo(b, off, f.lnext)
			}
			if err != nil {
				return None, path, false, err
			}

			if f.lastl == None {
				continue
			}

			f.lastp = append(f.lastp[:0], path[len(base):]...)
		} else if addpath {
			path = append(path[:len(base)], f.lastp...)
		}

		res, f.rnext, err = f.Index.ApplyTo(b, off, f.rnext)
		if err != nil {
			return None, path, false, err
		}

		if res == None {
			continue
		}

		break
	}

	if f.lastl == None || res == None {
		return None, path[:len(base)], false, nil
	}

	br := b.Reader()
	tag := br.Tag(res)

	switch tag {
	case cbor.Int, cbor.Neg:
		idx := br.Int(res)

		res, path, err = indexApplyTo(idx, b, f.lastl, path, false, addpath, null)
	case cbor.Bytes, cbor.String:
		key := br.Bytes(res)

		res, path, err = keyApplyTo(string(key), b, f.lastl, path, false, addpath, null, res)
	default:
		return None, base, false, NewTypeError(tag, cbor.Int, cbor.Neg, cbor.Bytes, cbor.String)
	}
	if err != nil {
		return None, path[:len(base)], false, err
	}

	more = f.lnext || f.rnext

	return res, path, more, nil
}

func (f ArbitraryIndex) String() string {
	return fmt.Sprintf(".(%v)[%v]%s", f.Base, f.Index, csel(f.NoErr, "?", ""))
}
