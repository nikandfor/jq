package jq

import "testing"

func TestPipe(tb *testing.T) {
	d, root := appendValBuf(nil, 0, obj{"a", obj{"b", obj{"c", "d"}}})
	b := NewBuffer(d)

	testOne(tb, NewPipe(Key("a"), Key("b"), Key("c")), b, root, "d")

	if tb.Failed() {
		tb.Logf("buffer\n%s", DumpBuffer(b))
	}

	d, root = appendValBuf(d, 0, "a")
	b.Reset(d)

	testIter(tb, NewPipe(
		NewComma(Dot{}, Dot{}),
		NewComma(Dot{}, Dot{}),
	), b, root, []any{"a", "a", "a", "a"})

	d, root = appendValBuf(d, 0, arr{arr{arr{"a", "b"}, arr{"c", "d"}}})
	b.Reset(d)

	testIter(tb, NewPipe(
		&Iter{}, &Iter{}, &Iter{},
	), b, root, []any{"a", "b", "c", "d"})
}
