package jq

import "testing"

func TestSlice(tb *testing.T) {
	b := NewBuffer(nil)

	root := b.appendVal(arr{0, 1, 2, 3, 4, 5})
	arr1 := b.appendVal(arr{1, 2, 3, 4})
	arr2 := b.appendVal(arr{5, 0})
	arr3 := b.appendVal(arr{})

	testOne(tb, NewSlice(0, 6), b, root, code(root))
	testOne(tb, NewSlice(-100, 100), b, root, code(root))
	testOne(tb, NewSlice(1, 5), b, root, code(arr1))
	testOne(tb, NewSlice(1, -1), b, root, code(arr1))
	testOne(tb, NewSlice(-5, -1), b, root, code(arr1))
	testOne(tb, NewSlice(-1, 1), b, root, code(arr2))
	testOne(tb, NewSlice(100, -100), b, root, code(arr3))
}