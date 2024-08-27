package jq

import (
	"testing"
)

func TestObject(tb *testing.T) {
	d, root := appendValBuf(nil, 0, obj{
		"a", "q",
		"b", 1,
		"c",
		obj{"d", "e"},
		"f",
		arr{2, 3},
	})

	b := NewBuffer(d)
	f := NewObject("a", NewIndex("a"), NewIndex("a"), NewIndex("f"))

	eoff := b.appendVal(obj{"a", "q", "q", arr{2, 3}})

	off, more, err := f.ApplyTo(b, root, false)
	assertNoError(tb, err)
	assertEqualVal(tb, b, eoff, off)
	assertTrue(tb, !more)

	// testIter(tb, f, b, root, []any{obj{"a", "q", "q", arr{2, 3}}})
}

func TestObjectIter(tb *testing.T) {
	d, root := appendValBuf(nil, 0, obj{
		"a", "b",
		"c",
		obj{
			"d", arr{
				obj{"e", "val1"},
				obj{"e", "val2"},
			},
		},
	})

	b := NewBuffer(d)
	f := NewObject("a", NewIndex("a"), "e", NewIndex("c", "d", Iter{}, "e"))

	testIter(tb, f, b, root, []any{
		obj{"a", "b", "e", "val1"},
		obj{"a", "b", "e", "val2"},
	})
}

func TestObjectIterMulti(tb *testing.T) {
	d, root := appendValBuf(nil, 0, obj{
		"a",
		arr{"q", "w", "e"},
		"c",
		obj{
			"d", arr{
				obj{"e", "val1"},
				obj{"e", "val2"},
			},
		},
	})

	b := NewBuffer(d)
	f := NewObject("a", NewIndex("a", Iter{}), "e", NewIndex("c", "d", Iter{}, "e"))

	testIter(tb, f, b, root, []any{
		obj{"a", "q", "e", "val1"},
		obj{"a", "q", "e", "val2"},
		obj{"a", "w", "e", "val1"},
		obj{"a", "w", "e", "val2"},
		obj{"a", "e", "e", "val1"},
		obj{"a", "e", "e", "val2"},
	})
}
