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

func (f Index) ApplyToGetPath(b *Buffer, base Path, at int, next bool) (res int, path Path, at1 int, more bool, err error) {
	return indexApplyToGetPath(int(f), b, base, at, next, false)
}

func (f IndexOrNull) ApplyToGetPath(b *Buffer, base Path, at int, next bool) (res int, path Path, at1 int, more bool, err error) {
	return indexApplyToGetPath(int(f), b, base, at, next, true)
}

func indexApplyToGetPath(f int, b *Buffer, base Path, at int, next, null bool) (res int, path Path, at1 int, more bool, err error) {
	off := base[at]

	res, more, err = indexApplyTo(int(f), b, off, next, null)
	if err != nil {
		return off, base, at, false, err
	}

	path = base
	at++

	return res, path, at, more, nil
}

func (f Index) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	return indexApplyTo(int(f), b, off, next, false)
}

func (f IndexOrNull) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	return indexApplyTo(int(f), b, off, next, true)
}

func indexApplyTo(f int, b *Buffer, off int, next, null bool) (res int, more bool, err error) {
	if next || off == None {
		return None, false, nil
	}
	if b.Equal(off, Null) {
		return Null, false, nil
	}

	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Array && tag != cbor.Map {
		if null {
			return Null, false, nil
		}

		return off, false, ErrType
	}

	l := br.ArrayMapLen(off)

	if f < 0 {
		f = l + f
	}

	if f < 0 || f >= l {
		return Null, false, nil
	}

	_, res = b.Reader().ArrayMapIndex(off, int(f))

	return res, false, nil
}

func (f Key) ApplyToGetPath(b *Buffer, base Path, at int, next bool) (res int, path Path, at1 int, more bool, err error) {
	return keyApplyToGetPath(string(f), b, base, at, next, false)
}

func (f KeyOrNull) ApplyToGetPath(b *Buffer, base Path, at int, next bool) (res int, path Path, at1 int, more bool, err error) {
	return keyApplyToGetPath(string(f), b, base, at, next, true)
}

func keyApplyToGetPath(f string, b *Buffer, base Path, at int, next, null bool) (res int, path Path, at1 int, more bool, err error) {
	off := base[at]

	res, more, err = keyApplyTo(f, b, off, next, null)
	if err != nil {
		return off, base, at, false, err
	}

	path = base
	at++

	return res, path, at, more, nil
}

func (f Key) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	return keyApplyTo(string(f), b, off, next, false)
}

func (f KeyOrNull) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	return keyApplyTo(string(f), b, off, next, true)
}

func keyApplyTo(f string, b *Buffer, off int, next, null bool) (res int, more bool, err error) {
	if next || off == None {
		return None, false, nil
	}
	if b.Equal(off, Null) {
		return Null, false, nil
	}

	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Map {
		if null {
			return Null, false, nil
		}

		return off, false, ErrType
	}

	l := br.ArrayMapLen(off)

	for j := range l {
		k, v := br.ArrayMapIndex(off, j)
		if string(br.Bytes(k)) == string(f) {
			return v, false, nil
		}
	}

	return Null, false, nil
}

func (f Index) String() string { return fmt.Sprintf(".[%d]", int(f)) }
func (f Key) String() string   { return fmt.Sprintf(".%s", string(f)) }
