package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Select struct {
		Cond Filter
	}
)

func NewSelect(cond Filter) *Select { return &Select{Cond: cond} }

func (f *Select) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if next {
		return None, false, nil
	}

	subf := f.Cond
	if subf == nil {
		subf = Dot{}
	}

	next = false

	for {
		res, next, err = subf.ApplyTo(b, off, next)
		if err != nil {
			return off, false, err
		}

		if f.IsTrue(b, res) {
			return off, false, nil
		}
		if !next {
			break
		}
	}

	return None, false, nil
}

func (f *Select) IsTrue(b *Buffer, off int) bool {
	if off < 0 {
		return off != None && off != Null && off != False
	}

	tag, sub, _, _, _ := b.Decoder.Tag(b.Buf(off))

	return tag != cbor.Simple || (sub != cbor.Null && sub != cbor.False && sub != cbor.None)
}

func (f Select) String() string {
	return fmt.Sprintf("Select(%v)", f.Cond)
}
