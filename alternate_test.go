package jq

import "testing"

func TestAlternate(tb *testing.T) {
	b := &Buffer{}
	d, root := appendValBuf(nil, 0, obj{
		"a",
		arr{},
		"b",
		arr{1, nil, false, 2},
		"c",
		arr{nil, false},
		"d", 5,
	})
	b.Reset(d)

	testIter(tb, NewAlternate(NewQuery("a", Iter{}), NewQuery("b", Iter{})), b, root, []any{1, nil, false, 2})
	testIter(tb, NewAlternate(NewQuery("b", Iter{}), NewQuery("d")), b, root, []any{1, 2})
	testIter(tb, NewAlternate(NewQuery("c", Iter{}), NewQuery("d")), b, root, []any{5})
}
