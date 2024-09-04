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

func (f Index) ApplyToGetPath(b *Buffer, off int, next bool, base Path) (res int, path Path, more bool, err error) {
	return indexApplyToGetPath(int(f), b, off, next, false, base)
}

func (f IndexOrNull) ApplyToGetPath(b *Buffer, off int, next bool, base Path) (res int, path Path, more bool, err error) {
	return indexApplyToGetPath(int(f), b, off, next, true, base)
}

func indexApplyToGetPath(f int, b *Buffer, off int, next, null bool, base Path) (res int, path Path, more bool, err error) {
	res, more, err = indexApplyTo(int(f), b, off, next, null)
	if err != nil {
		return off, base, false, err
	}

	path = append(base, off)

	return res, path, more, nil
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

func (f Key) ApplyToGetPath(b *Buffer, off int, next bool, base Path) (res int, path Path, more bool, err error) {
	return keyApplyToGetPath(string(f), b, off, next, false, base)
}

func (f KeyOrNull) ApplyToGetPath(b *Buffer, off int, next bool, base Path) (res int, path Path, more bool, err error) {
	return keyApplyToGetPath(string(f), b, off, next, true, base)
}

func keyApplyToGetPath(f string, b *Buffer, off int, next, null bool, base Path) (res int, path Path, more bool, err error) {
	res, more, err = keyApplyTo(f, b, off, next, null)
	if err != nil {
		return off, base, false, err
	}

	path = append(base, off)

	return res, path, more, nil
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
