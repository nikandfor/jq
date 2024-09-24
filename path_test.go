package jq

import "testing"

func TestPath(tb *testing.T) {
	b := NewBuffer()
	r := b.appendVal(obj{"a", arr{obj{"b", 1}, obj{"b", 2}}})

	testIter(tb, NewPath(NewQuery("a", Iter{}, "b")), b, r, []any{
		arr{"a", 0, "b"},
		arr{"a", 1, "b"},
	})
}
