package jq

import (
	"runtime"
	"testing"
)

func TestComma(tb *testing.T) {
	d, root := appendValBuf(nil, arr{4, 3, 2, 1})
	b := NewBuffer(d)

	testOne(tb, NewComma(), b, root, code(None))
	testIter(tb, NewComma(NewIndex(3), NewIndex(2), NewIndex(1), NewIndex(0)), b, root, []any{1, 2, 3, 4})

	if tb.Failed() {
		return
	}

	d, root = appendValBuf(d, arr{arr{3, 4}, arr{1, 2}})
	b.Reset(d)

	testIter(tb, NewComma(NewIndex(1, Iter{}), NewIndex(0, Iter{})), b, root, []any{1, 2, 3, 4})

	// tb.Logf("buffer\n%s", DumpBuffer(b))
}

func testOne(tb testing.TB, f Filter, b *Buffer, root int, val any) {
	eoff := b.appendVal(val)

	off, more, err := f.ApplyTo(b, root, false)
	assertNoError(tb, err)
	assertEqualVal(tb, b, eoff, off, "wanted %v", val)
	assertTrue(tb, !more, "didn't want more")

	if tb.Failed() {
		_, file, line, _ := runtime.Caller(1)
		tb.Logf("from %v:%d", file, line)
	}
}
