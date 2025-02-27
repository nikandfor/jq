package jq

import (
	"testing"
)

func TestSelect(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(arr{
		nil,
		false,
		true,
		0,
		1,
		"a",
		obj{},
		arr{},
		lab{lab: 10, val: 1},
		0,
	})

	testOne(tb, NewArray(NewPipe(NewIter(), NewSelect(nil))), b, root, arr{
		true,
		0,
		1,
		"a",
		obj{},
		arr{},
		lab{lab: 10, val: 1},
		0,
	})

	b.Reset()
	root = b.appendVal(arr{
		obj{"a", nil},
		obj{"a", false},
		obj{"a", true},
		obj{"a", 0},
		obj{"a", 1},
		obj{"a", "a"},
		obj{"a", obj{}},
		obj{"a", arr{}},
		obj{"a", lab{lab: 10, val: 1}},
	})

	testOne(tb, NewArray(NewPipe(NewIter(), NewSelect(NewQuery("a")))), b, root, arr{
		obj{"a", true},
		obj{"a", 0},
		obj{"a", 1},
		obj{"a", "a"},
		obj{"a", obj{}},
		obj{"a", arr{}},
		obj{"a", lab{lab: 10, val: 1}},
	})
}

func TestSelectSpecial(tb *testing.T) {
	b := NewBuffer()
	_ = b.appendVal(make([]byte, 250))

	root := b.appendVal(arr{
		Null,
		True,
		False,
		Zero,
		One,
		obj{},
	})

	testOne(tb, NewArray(NewPipe(NewIter(), NewSelect(nil))), b, root, arr{true, 0, 1, obj{}})
}

func TestPipeSelect(tb *testing.T) {
	b := NewBuffer()
	off := b.appendVal(arr{obj{"a", false, "v", 1}, obj{"a", true, "v", 2}, obj{"a", 1, "v", 3}, obj{"a", nil, "v", 4}})
	exp := b.appendVal(arr{obj{"a", true, "v", 2}, obj{"a", 1, "v", 3}})

	testOne(tb, NewArray(NewPipe(NewIter(), NewSelect(Key("a")))), b, off, exp)
}
