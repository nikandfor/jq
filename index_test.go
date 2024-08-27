package jq

import (
	"testing"
)

func TestIndex(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", 1, "b", obj{"c", arr{2, "3", obj{"d", 5}, true}}})

	//	log.Printf("data %x\n%s", root, Dump(d))

	f := NewIndex("a")
	eoff := b.appendVal(1)

	off, more, err := f.ApplyTo(b, root, false)
	assertNoError(tb, err)
	assertEqualVal(tb, b, eoff, off)
	assertTrue(tb, !more)

	f = NewIndex("b", "c")
	eoff = b.appendVal(arr{2, "3", obj{"d", 5}, true})

	off, more, err = f.ApplyTo(b, root, false)
	assertNoError(tb, err)
	assertEqualVal(tb, b, eoff, off)
	assertTrue(tb, !more)

	f = NewIndex("b", "c", Iter{})
	testIter(tb, f, b, root, []any{2, "3", obj{"d", 5}, true})
}

func TestIndexIter(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(arr{
		obj{"a", 1, "b", lab{lab: 4, val: 2}, "c", "d"},
		true,
	})

	//	log.Printf("data %x\n%s", root, Dump(d))

	f := NewIndex(-2, Iter{})
	testIter(tb, f, b, root, []any{1, lab{lab: 4, val: 2}, "d"})
}

func TestIndexMultiIter(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(arr{
		obj{"q", obj{"a", 1, "b", 2}},
		obj{"q", arr{}, "w", -5},
		obj{"q", arr{3, 4}},
	})

	//	log.Printf("data %x\n%s", root, Dump(d))

	f := NewIndex(Iter{}, "q", Iter{})

	testIter(tb, f, b, root, []any{1, 2, 3, 4})
}

func TestIndexIgnoreTypeError(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", "b"})

	testOne(tb, NewIndex("a"), b, root, "b")
	testOne(tb, NewIndex("q"), b, root, nil)

	root = b.appendVal(arr{"a", "b"})

	f := NewIndex("a")
	f.IgnoreTypeError = true

	testOne(tb, f, b, root, nil)

	f = NewIndex("a")
	f.IgnoreTypeError = false

	testError(tb, f, b, root, ErrType)
}
