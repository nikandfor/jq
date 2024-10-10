package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Index        int
	IndexNoError int

	Key        string
	KeyNoError string

	addpath bool
)

const (
	withPath    addpath = true
	withoutPath addpath = false
)

var (
	_ FilterPath = Index(0)
	_ FilterPath = IndexNoError(0)

	_ FilterPath = Key("")
	_ FilterPath = KeyNoError("")
)

func (f Index) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	res, path, err = indexApplyTo(int(f), b, off, base, next, withPath, false)
	return
}

func (f IndexNoError) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	res, path, err = indexApplyTo(int(f), b, off, base, next, withPath, true)
	return
}

func (f Index) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, err = indexApplyTo(int(f), b, off, nil, next, withoutPath, false)
	return
}

func (f IndexNoError) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, err = indexApplyTo(int(f), b, off, nil, next, withoutPath, true)
	return
}

func indexApplyTo(f int, b *Buffer, off Off, base NodePath, next bool, addpath addpath, noerr bool) (res Off, path NodePath, err error) {
	if off == None || next {
		return None, base, nil
	}

	path = base

	if addpath {
		path = append(base, NodePathSeg{Off: off, Index: f, Key: None})
	}

	if b.Equal(off, Null) {
		return Null, path, nil
	}

	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Array && tag != cbor.Map {
		if noerr {
			return None, path, nil
		}

		return off, path, NewTypeError(tag, cbor.Array, cbor.Map)
	}

	l := br.ArrayMapLen(off)

	if f < 0 {
		f = l + f
	}

	if f < 0 || f >= l {
		return Null, path, nil
	}

	_, res = b.Reader().ArrayMapIndex(off, f)

	if addpath {
		path[len(path)-1].Index = f
	}

	return res, path, nil
}

func (f Key) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	res, path, err = keyApplyTo(string(f), b, off, base, next, true, false, None)
	return
}

func (f KeyNoError) ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	res, path, err = keyApplyTo(string(f), b, off, base, next, true, true, None)
	return
}

func (f Key) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, err = keyApplyTo(string(f), b, off, nil, next, false, false, None)
	return
}

func (f KeyNoError) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, _, err = keyApplyTo(string(f), b, off, nil, next, false, true, None)
	return
}

func keyApplyTo(f string, b *Buffer, off Off, base NodePath, next bool, addpath addpath, noerr bool, keyoff Off) (res Off, path NodePath, err error) {
	if off == None || next {
		return None, base, nil
	}

	path = base

	if addpath {
		path = append(base, NodePathSeg{Off: off, Index: -1, Key: keyoff})
	}

	br := b.Reader()
	var l int

	if !b.Equal(off, Null) {
		tag := br.Tag(off)
		if tag != cbor.Map {
			if noerr {
				return None, path[:len(base)], nil
			}

			return None, path, NewTypeError(tag, cbor.Map)
		}

		l = br.ArrayMapLen(off)
	}

	for j := range l {
		k, v := br.ArrayMapIndex(off, j)
		if string(br.Bytes(k)) != f {
			continue
		}

		if addpath {
			last := len(path) - 1
			path[last].Index = j
			path[last].Key = k
		}

		return v, path, nil
	}

	if addpath {
		if keyoff == None {
			keyoff = b.Writer().String(f)
		}

		path[len(path)-1].Key = keyoff
	}

	return Null, path, nil
}

func (f Index) String() string { return fmt.Sprintf(".[%d]", int(f)) }
func (f Key) String() string   { return fmt.Sprintf(".%s", string(f)) }
