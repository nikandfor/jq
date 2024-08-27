package jq

import "testing"

func TestIter(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", 1, "b", 2, "c", "d"})

	testIter(tb, NewIter(), b, root, []any{1, 2, "d"})
}
