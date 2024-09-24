package jq

import (
	"testing"
)

func TestComma(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(arr{4, 3, 2, 1})

	testOne(tb, NewComma(), b, root, None)
	testIter(tb, NewComma(NewQuery(3), NewQuery(2), NewQuery(1), NewQuery(0)), b, root, []any{1, 2, 3, 4})

	if tb.Failed() {
		return
	}

	b.Reset()
	root = b.appendVal(arr{arr{3, 4}, arr{1, 2}})

	testIter(tb, NewComma(NewQuery(1, Iter{}), NewQuery(0, Iter{})), b, root, []any{1, 2, 3, 4})

	// tb.Logf("buffer\n%s", DumpBuffer(b))
}
