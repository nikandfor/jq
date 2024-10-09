package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestQuery(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", 1, "b", obj{"c", arr{2, "3", obj{"d", 5}, true}}})

	testOne(tb, NewQuery("a"), b, root, 1)

	if tb.Failed() {
		tb.Logf("buffer  root %x\n%s", root, b.Dump())
		return
	}

	testOne(tb, NewQuery("b", "c"), b, root, arr{2, "3", obj{"d", 5}, true})
}

func TestQueryIter1(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", 1, "b", obj{"c", arr{2, "3", obj{"d", 5}, true}}})

	testIter(tb, NewQuery("b", "c", Iter{}), b, root, []any{2, "3", obj{"d", 5}, true})
}

func TestQueryIter2(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(arr{
		obj{"a", 1, "b", lab{lab: 4, val: 2}, "c", "d"},
		true,
	})

	//	log.Printf("data %x\n%s", root, Dump(d))

	testIter(tb, NewQuery(-2, Iter{}), b, root, []any{1, lab{lab: 4, val: 2}, "d"})
}

func TestQueryMultiIter(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(arr{
		obj{"q", obj{"a", 1, "b", 2}},
		obj{"q", arr{}, "w", -5},
		obj{"q", arr{3, 4}},
	})

	//	log.Printf("data %x\n%s", root, Dump(d))

	f := NewQuery(Iter{}, "q", Iter{})

	testIter(tb, f, b, root, []any{1, 2, 3, 4})
}

func TestQueryIgnoreTypeError(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", "b"})

	testOne(tb, NewQuery("a"), b, root, "b")
	testOne(tb, NewQuery("q"), b, root, nil)

	root2 := b.appendVal(arr{"a", "b"})

	testIter(tb, NewQuery(KeyNoErr("a")), b, root2, []any{})
	testError(tb, NewQuery("a"), b, root2, NewTypeError(cbor.Array, cbor.Map))

	if tb.Failed() {
		tb.Logf("buffer\n%s", b.Dump())
	}
}
