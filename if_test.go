package jq

import "testing"

func TestIfPipe(tb *testing.T) {
	b := NewBuffer()
	off := b.appendVal(arr{false, 1, 2})

	testIter(tb, NewPipe(NewIter(), NewIf(nil, NewObject("x", Dot{}), NewLiteral("no"))), b, off, []any{
		"no", obj{"x", 1}, obj{"x", 2},
	})
}

func TestIfIter(tb *testing.T) {
	b := NewBuffer()
	off := b.appendVal(arr{false, 1, 2})

	testIter(tb, NewIf(NewIter(), NewObject("x", Dot{}), NewLiteral("no")), b, off, []any{
		"no", obj{"x", arr{false, 1, 2}}, obj{"x", arr{false, 1, 2}},
	})
}
