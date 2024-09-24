package jq

import "testing"

func TestArray(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(arr{
		obj{"a", 1},
		obj{"a", "2"},
		obj{"a", 3},
		obj{"a", "4"},
	})

	f := NewArray(NewQuery(Iter{}, "a"))

	testOne(tb, f, b, root, arr{1, "2", 3, "4"})

	if tb.Failed() {
		tb.Logf("buffer\n%s", Dump(b))
	}
}
