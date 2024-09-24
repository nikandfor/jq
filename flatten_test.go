package jq

import "testing"

func TestFlatten(tb *testing.T) {
	b := NewBuffer()

	d6 := b.appendVal(arr{6, 7})
	d5 := b.appendVal(arr{"5", d6})
	d4 := b.appendVal(arr{obj{"a", 4}, d5})
	d3 := b.appendVal(arr{3, d4})
	d2 := b.appendVal(arr{2, d3})
	d1 := b.appendVal(arr{1, d2})

	root := d1

	testSame(tb, NewFlatten(0), b, root, root)
	testOne(tb, NewFlatten(1), b, root, arr{1, 2, d3})
	testOne(tb, NewFlatten(2), b, root, arr{1, 2, 3, d4})
	testOne(tb, NewFlatten(3), b, root, arr{1, 2, 3, obj{"a", 4}, d5})
	testOne(tb, NewFlatten(4), b, root, arr{1, 2, 3, obj{"a", 4}, "5", d6})
	testOne(tb, NewFlatten(5), b, root, arr{1, 2, 3, obj{"a", 4}, "5", 6, 7})
	testOne(tb, NewFlatten(-1), b, root, arr{1, 2, 3, obj{"a", 4}, "5", 6, 7})
}
