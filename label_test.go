package jq

import "testing"

func TestLabel(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		1,
		"str",
		lab{lab: 0, val: 0},
		lab{lab: 10, val: 10},
	})

	testIter(tb, NewPipe(&Iter{}, Label{}), b, root, []any{
		Null,
		Null,
		0,
		10,
	})
}
