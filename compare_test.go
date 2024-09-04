package jq

import (
	"testing"
)

func TestCompareEqualBasic(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", obj{"b", 10}, "b", 10})
	root2 := b.appendVal(obj{"a", obj{"b", 10}, "b", 20})

	testOne(tb, NewEqual(NewIndex("a", "b"), NewIndex("b")), b, root, code(True))
	testOne(tb, NewEqual(NewIndex("a", "b"), NewIndex("b")), b, root2, code(False))

	testOne(tb, NewNotEqual(NewIndex("a", "b"), NewIndex("b")), b, root, code(False))
	testOne(tb, NewNotEqual(NewIndex("a", "b"), NewIndex("b")), b, root2, code(True))
}

func TestCompareEqualMultiOne(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", arr{1, 2, 3}, "b", arr{1, 2, 3}})

	testIter(tb, NewEqual(NewIndex("a", Iter{}), Off(One)), b, root, []any{
		code(True), code(False), code(False),
	})
}

func TestCompareEqualMultiNone(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", arr{1, 2, 3}, "b", arr{}})

	testIter(tb, NewEqual(NewIndex("a", Iter{}), NewIndex("b", Iter{})), b, root, []any{})
}

func TestCompareEqualMultiMulti(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", arr{1, 2, 3}, "b", arr{1, 2, 3}})

	testIter(tb, NewEqual(NewIndex("a", Iter{}), NewIndex("b", Iter{})), b, root, []any{
		true, false, false,
		false, true, false,
		false, false, true,
	})

	if tb.Failed() {
		tb.Logf("buffer  %x\n%s", root, DumpBuffer(b))
	}
}

func TestCompareNot(tb *testing.T) {
	b := NewBuffer(nil)

	testOne(tb, NewNot(), b, b.appendVal(true), false)
	testOne(tb, NewNot(), b, b.appendVal(1), false)
	testOne(tb, NewNot(), b, b.appendVal(arr{}), false)
	testOne(tb, NewNot(), b, b.appendVal(obj{}), false)
	testOne(tb, NewNot(), b, b.appendVal(false), true)
	testOne(tb, NewNot(), b, b.appendVal(nil), true)
	testOne(tb, NewNot(), b, b.appendVal(code(None)), code(None))
}

func TestCompareNotOf(tb *testing.T) {
	b := NewBuffer(nil)

	testOne(tb, NewNotOf(Dot{}), b, b.appendVal(true), false)
	testOne(tb, NewNotOf(NewNotOf(Dot{})), b, b.appendVal(1), true)
}
