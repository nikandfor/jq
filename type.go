package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	TypeMatch struct {
		TypeSet TypeSet
	}

	TypeSet int64
)

func NewTypeMatch(tags ...Tag) TypeMatch {
	return TypeMatch{TypeSet: NewTypeSet(tags...)}
}

func (f TypeMatch) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	tag := b.Reader().TagRaw(off)

	if f.TypeSet.Match(tag) {
		return True, false, nil
	}

	return False, false, nil
}

func NewTypeSet(tags ...Tag) TypeSet {
	var ts TypeSet

	for i, t := range tags {
		ts |= 1 << (t >> 5)

		if t&cbor.TagMask == cbor.Simple && i < 6 {
			ts |= TypeSet(t) << (8 * (1 + i))
		}
	}

	return ts
}

func (ts TypeSet) Match(tag Tag) bool {
	t := tag >> 5

	if ts&(1<<t) == 0 {
		return false
	}
	if tag&cbor.TagMask != cbor.Simple {
		return true
	}

	for i := range 6 {
		if Tag(ts>>(8*(1+i))) == tag {
			return true
		}
	}

	return false
}

func (f TypeMatch) String() string {
	return fmt.Sprintf("TypeMatch(%#x)", int64(f.TypeSet))
}
