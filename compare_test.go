package jq

import (
	"testing"
)

func TestCompareEqualBasic(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", obj{"b", 10}, "b", 10})
	root2 := b.appendVal(obj{"a", obj{"b", 10}, "b", 20})

	testOne(tb, NewEqual(NewQuery("a", "b"), NewQuery("b")), b, root, True)
	testOne(tb, NewEqual(NewQuery("a", "b"), NewQuery("b")), b, root2, False)

	testOne(tb, NewNotEqual(NewQuery("a", "b"), NewQuery("b")), b, root, False)
	testOne(tb, NewNotEqual(NewQuery("a", "b"), NewQuery("b")), b, root2, True)
}

func TestCompareEqualMultiOne(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", arr{1, 2, 3}, "b", arr{1, 2, 3}})

	testIter(tb, NewEqual(NewQuery("a", Iter{}), Off(One)), b, root, []any{
		True, False, False,
	})
}

func TestCompareEqualMultiNone(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", arr{1, 2, 3}, "b", arr{}})

	testIter(tb, NewEqual(NewQuery("a", Iter{}), NewQuery("b", Iter{})), b, root, []any{})
}

func TestCompareEqualMultiMulti(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", arr{1, 2, 3}, "b", arr{1, 2, 3}})

	testIter(tb, NewEqual(NewQuery("a", Iter{}), NewQuery("b", Iter{})), b, root, []any{
		true, false, false,
		false, true, false,
		false, false, true,
	})

	if tb.Failed() {
		tb.Logf("buffer  %x\n%s", root, b.Dump())
	}
}

func TestCompareNot(tb *testing.T) {
	b := NewBuffer()

	testOne(tb, NewNot(), b, b.appendVal(true), false)
	testOne(tb, NewNot(), b, b.appendVal(1), false)
	testOne(tb, NewNot(), b, b.appendVal(arr{}), false)
	testOne(tb, NewNot(), b, b.appendVal(obj{}), false)
	testOne(tb, NewNot(), b, b.appendVal(false), true)
	testOne(tb, NewNot(), b, b.appendVal(nil), true)
	testOne(tb, NewNot(), b, b.appendVal(None), None)
}

func TestCompareNotOf(tb *testing.T) {
	b := NewBuffer()

	testOne(tb, NewNotOf(Dot{}), b, b.appendVal(true), false)
	testOne(tb, NewNotOf(NewNotOf(Dot{})), b, b.appendVal(1), true)
}
