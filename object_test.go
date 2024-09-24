package jq

import (
	"testing"
)

func TestObject(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{
		"a", "q",
		"b", 1,
		"c",
		obj{"d", "e"},
		"f",
		arr{2, 3},
	})

	f := NewObject("a", NewQuery("a"), NewQuery("a"), NewQuery("f"))

	testOne(tb, f, b, root, obj{"a", "q", "q", arr{2, 3}})
}

func TestObjectIter(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{
		"a", "b",
		"c",
		obj{
			"d", arr{
				obj{"e", "val1"},
				obj{"e", "val2"},
			},
		},
	})

	f := NewObject("a", NewQuery("a"), "e", NewQuery("c", "d", Iter{}, "e"))

	testIter(tb, f, b, root, []any{
		obj{"a", "b", "e", "val1"},
		obj{"a", "b", "e", "val2"},
	})
}

func TestObjectIterMulti(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{
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

	f := NewObject("a", NewQuery("a", Iter{}), "e", NewQuery("c", "d", Iter{}, "e"))

	testIter(tb, f, b, root, []any{
		obj{"a", "q", "e", "val1"},
		obj{"a", "q", "e", "val2"},
		obj{"a", "w", "e", "val1"},
		obj{"a", "w", "e", "val2"},
		obj{"a", "e", "e", "val1"},
		obj{"a", "e", "e", "val2"},
	})
}

func TestObjectCopyKeys(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", 1, "b", 2})

	f := NewObject(
		ObjectCopyKey("a"),
		ObjectCopyKey("b"),
		ObjectCopyKey("c"),
	)

	testOne(tb, f, b, root, obj{"a", 1, "b", 2, "c", nil})
}
