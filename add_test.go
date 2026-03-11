package jq

import "testing"

func TestAddNumbers(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		Zero, One, 2, 3, -5,
	})

	testOne(tb, NewAdd(), b, root, One)

	root = b.appendVal(arr{
		1, 1.5, -3.5,
	})

	testOne(tb, NewAdd(), b, root, -1.)
}

func TestAddGenNumbers(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		Zero, One, 2, 3, -5,
	})

	testOne(tb, NewAddGen(NewIter()), b, root, One)

	root = b.appendVal(arr{
		1, 1.5, -3.5,
	})

	testOne(tb, NewAddGen(NewIter()), b, root, -1.)
}

func TestAddStrings(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		"aa", "bbb",
	})

	testOne(tb, NewAdd(), b, root, "aabbb")
}

func TestAddGenStrings(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		"aa", "bbb",
	})

	testOne(tb, NewAddGen(NewIter()), b, root, "aabbb")
}

func TestAddArray(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		arr{"aa"}, arr{"bbb"},
	})

	testOne(tb, NewAdd(), b, root, arr{"aa", "bbb"})
}

func TestAddGenArray(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		arr{"aa"}, arr{"bbb"},
	})

	testOne(tb, NewAddGen(NewIter()), b, root, arr{"aa", "bbb"})
}

func TestAddObject(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		obj{"a", 1}, obj{"b", 2},
	})

	testOne(tb, NewAdd(), b, root, obj{"a", 1, "b", 2})
}

func TestAddGenObject(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		obj{"a", 1}, obj{"b", 2},
	})

	testOne(tb, NewAddGen(NewIter()), b, root, obj{"a", 1, "b", 2})
}

func TestAddMaps(tb *testing.T) {
	b := NewBuffer()

	q := b.appendVal(obj{"q", 1})
	w := b.appendVal(obj{"w", 2})
	e := b.appendVal(obj{"e", 3})

	sum := b.appendVal(obj{"q", 1, "w", 2, "e", 3})

	res, err := AddMaps(b, q, w, e)
	assertNoError(tb, err)
	assertEqualVal(tb, b, sum, res)
}
