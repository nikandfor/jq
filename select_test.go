package jq

import (
	"testing"
)

func TestSelect(tb *testing.T) {
	d, root := appendValBuf(nil, 0, arr{
		nil,
		false,
		true,
		0,
		1,
		"a",
		obj{},
		arr{},
	})
	b := NewBuffer(d)

	testOne(tb, NewArray(NewPipe(NewIter(), NewSelect(nil))), b, root, arr{true, 0, 1, "a", obj{}, arr{}})

	d, root = appendValBuf(d, 0, arr{
		obj{"a", nil},
		obj{"a", false},
		obj{"a", true},
		obj{"a", 0},
		obj{"a", 1},
		obj{"a", "a"},
		obj{"a", obj{}},
		obj{"a", arr{}},
	})
	b.Reset(d)

	testOne(tb, NewArray(NewPipe(NewIter(), NewSelect(NewQuery("a")))), b, root, arr{
		obj{"a", true},
		obj{"a", 0},
		obj{"a", 1},
		obj{"a", "a"},
		obj{"a", obj{}},
		obj{"a", arr{}},
	})
}

func TestSelectSpecial(tb *testing.T) {
	b := NewBuffer(nil)
	_ = b.appendVal(make([]byte, 250))

	root := b.appendVal(arr{
		code(Null),
		code(True),
		code(False),
		code(Zero),
		code(One),
		obj{},
	})

	testOne(tb, NewArray(NewPipe(NewIter(), NewSelect(nil))), b, root, arr{true, 0, 1, obj{}})
}
