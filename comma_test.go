package jq

import (
	"testing"
)

func TestComma(tb *testing.T) {
	d, root := appendValBuf(nil, 0, arr{4, 3, 2, 1})
	b := NewBuffer(d)

	testOne(tb, NewComma(), b, root, code(None))
	testIter(tb, NewComma(NewIndex(3), NewIndex(2), NewIndex(1), NewIndex(0)), b, root, []any{1, 2, 3, 4})

	if tb.Failed() {
		return
	}

	d, root = appendValBuf(d, 0, arr{arr{3, 4}, arr{1, 2}})
	b.Reset(d)

	testIter(tb, NewComma(NewIndex(1, Iter{}), NewIndex(0, Iter{})), b, root, []any{1, 2, 3, 4})

	// tb.Logf("buffer\n%s", DumpBuffer(b))
}
