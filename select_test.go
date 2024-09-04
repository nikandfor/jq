package jq

import (
	"testing"

	"nikand.dev/go/cbor"
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
	d, root := appendValBuf(append([]byte{cbor.String | cbor.Len1, 250}, make([]byte, 250)...), 0, arr{
		code(None),
		code(Null),
		code(True),
		code(False),
		code(Zero),
		code(One),
		obj{},
	})
	b := NewBuffer(d)

	testOne(tb, NewArray(NewPipe(NewIter(), NewSelect(nil))), b, root, arr{true, 0, 1, obj{}})
}
