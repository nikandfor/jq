package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Select struct {
		Condition Filter
	}
)

func NewSelect(cond Filter) Select { return Select{Condition: cond} }

func (f Select) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	if f.Condition == nil {
		if IsTrue(b, off) {
			return off, false, nil
		}

		return None, false, nil
	}

	next = false

	for {
		res, next, err = f.Condition.ApplyTo(b, off, next)
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

func ToBool(b *Buffer, off Off) Off {
	if IsTrue(b, off) {
		return True
	}

	return False
}

func IsTrue(b *Buffer, off Off) bool {
	if off < 0 {
		return off != None && off != Null && off != False
	}

	off = b.Reader().UnderAllLabels(off)
	tag, sub, _, _, _, _ := b.Decoder.Tag(b.Buf(off))

	return tag != cbor.Simple || (sub != cbor.Null && sub != cbor.False && sub != cbor.None)
}

func (f Select) String() string {
	return fmt.Sprintf("select(%v)", or[Filter](f.Condition, Dot{}))
}
