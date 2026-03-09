package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	HasIndex int
	HasKey   string

	Has struct {
		Filter Filter
	}
)

func NewHas(f any) Filter {
	switch f := f.(type) {
	case int:
		return HasIndex(f)
	case string:
		return HasKey(f)
	case Filter:
		return Has{Filter: f}
	default:
		panic(f)
	}
}

func (f Has) ApplyTo(b *Buffer, off Off, next bool) (idx Off, more bool, err error) {
	idx, more, err = f.Filter.ApplyTo(b, off, next)
	if err != nil || idx == None {
		return idx, more, err
	}

	br := b.Reader()

	switch tag := br.Tag(idx); true {
	case cbor.IsInt(tag):
		idx := br.Int(idx)
		l := br.ArrayMapLen(off)

		if idx >= 0 && idx < l || idx < 0 && l+idx >= 0 {
			return True, false, nil
		}

		return False, false, nil
	case tag == TagString:
		val := br.MapValue(off, idx)

		if val != None {
			return True, false, nil
		}

		return False, false, nil
	default:
		return off, false, NewTypeError(tag, TagInt, TagNeg, TagString)
	}
}

func (f HasIndex) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()

	if tag := br.Tag(off); tag != TagArray {
		return off, false, NewTypeError(tag, TagArray)
	}

	l := br.ArrayMapLen(off)

	if f >= 0 && int(f) < l || f < 0 && l+int(f) >= 0 {
		return True, false, nil
	}

	return False, false, nil
}

func (f HasKey) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()

	if tag := br.Tag(off); tag != TagMap {
		return off, false, NewTypeError(tag, TagMap)
	}

	val := br.MapValueStringKey(off, string(f))

	if val != None {
		return True, false, nil
	}

	return False, false, nil
}

func (f HasIndex) String() string {
	return fmt.Sprintf("has(%d)", int(f))
}

func (f HasKey) String() string {
	return fmt.Sprintf("has(%q)", string(f))
}

func (f Has) String() string {
	return fmt.Sprintf("has(%v)", f.Filter)
}
