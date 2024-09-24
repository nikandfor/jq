package jq

import "testing"

func TestAnyAll(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		arr{},
		arr{false, nil},
		arr{false, nil, 1},
		arr{1, 2, 3},
		obj{},
		obj{"a", false, "b", nil},
		obj{"a", false, "b", nil, "c", 1},
		obj{"a", 1},
	})

	testIter(tb, NewPipe(NewIter(), NewAny()), b, root, []any{
		false, false, true, true,
		false, false, true, true,
	})

	testIter(tb, NewPipe(NewIter(), NewAll()), b, root, []any{
		true, false, false, true,
		true, false, false, true,
	})
}
