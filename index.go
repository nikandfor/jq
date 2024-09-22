package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Index       int
	IndexOrNull int

	Key       string
	KeyOrNull string
)

var (
	_ FilterPath = Index(0)
	_ FilterPath = IndexOrNull(0)

	_ FilterPath = Key("")
	_ FilterPath = KeyOrNull("")
)

func (f Index) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	return indexApplyTo(int(f), b, off, base, next, true, false)
}

func (f IndexOrNull) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	return indexApplyTo(int(f), b, off, base, next, true, true)
}

func (f Index) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, more, err = indexApplyTo(int(f), b, off, nil, next, false, false)
	return
}

func (f IndexOrNull) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, more, err = indexApplyTo(int(f), b, off, nil, next, false, true)
	return
}

func indexApplyTo(f int, b *Buffer, off Off, base NodePath, next, addpath, null bool) (res Off, path NodePath, more bool, err error) {
	if next || off == None {
		return None, base, false, nil
	}
	if b.Equal(off, Null) {
		return Null, base, false, nil
	}

	if addpath {
		path = append(base, NodePathSeg{Off: off, Index: -1})
	}

	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Array && tag != cbor.Map {
		if null {
			return Null, path, false, nil
		}

		return off, path, false, NewTypeError(tag, cbor.Array, cbor.Map)
	}

	l := br.ArrayMapLen(off)

	if f < 0 {
		f = l + f
	}

	if f < 0 || f >= l {
		return Null, path, false, nil
	}

	_, res = b.Reader().ArrayMapIndex(off, f)

	if addpath {
		path[len(path)-1].Index = f
	}

	return res, path, false, nil
}

func (f Key) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	return keyApplyTo(string(f), b, off, base, next, true, false)
}

func (f KeyOrNull) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	return keyApplyTo(string(f), b, off, base, next, true, true)
}

func (f Key) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, more, err = keyApplyTo(string(f), b, off, nil, next, false, false)
	return
}

func (f KeyOrNull) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, more, err = keyApplyTo(string(f), b, off, nil, next, false, true)
	return
}

func keyApplyTo(f string, b *Buffer, off Off, base NodePath, next, addpath, null bool) (res Off, path NodePath, more bool, err error) {
	if next || off == None {
		return None, base, false, nil
	}
	if b.Equal(off, Null) {
		return Null, base, false, nil
	}

	if addpath {
		path = append(base, NodePathSeg{Off: off, Index: -1})
	}

	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Map {
		if null {
			return Null, path, false, nil
		}

		return off, path, false, NewTypeError(tag, cbor.Map)
	}

	l := br.ArrayMapLen(off)

	for j := range l {
		k, v := br.ArrayMapIndex(off, j)
		if string(br.Bytes(k)) != string(f) {
			continue
		}

		if addpath {
			path[len(path)-1].Index = j
		}

		return v, path, false, nil
	}

	return Null, path, false, nil
}

func (f Index) String() string { return fmt.Sprintf(".[%d]", int(f)) }
func (f Key) String() string   { return fmt.Sprintf(".%s", string(f)) }
