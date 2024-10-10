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

func (f *Select) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	subf := or[Filter](f.Cond, Dot{})

	next = false

	for {
		res, next, err = subf.ApplyTo(b, off, next)
		if err != nil {
			return off, false, fe(f, off, err)
		}

		if IsTrue(b, res) {
			return off, false, nil
		}
		if !next {
			break
		}
	}

	return None, false, nil
}

func IsTrue(b *Buffer, off Off) bool {
	if off < 0 {
		return off != None && off != Null && off != False
	}

	tag, sub, _, _, _ := b.Decoder.Tag(b.Buf(off))

	return tag != cbor.Simple || (sub != cbor.Null && sub != cbor.False && sub != cbor.None)
}

func (f Select) String() string {
	return fmt.Sprintf("select(%v)", or[Filter](f.Cond, Dot{}))
}
