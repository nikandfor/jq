package jq

import "testing"

func TestArray(tb *testing.T) {
	d, root := appendValBuf(nil, arr{
		obj{"a", 1},
		obj{"a", "2"},
		obj{"a", 3},
		obj{"a", "4"},
	})

	b := NewBuffer(d)
	f := NewArray(NewIndex(Iter{}, "a"))

	testOne(tb, f, b, root, arr{1, "2", 3, "4"})

	if tb.Failed() {
		tb.Logf("buffer\n%s", DumpBuffer(b))
	}
}
