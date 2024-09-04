//go:build ignore

package jq

import "testing"

func TestAssignObject(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", obj{"b", 10}, "b", 20})

	testOne(tb, NewAssign(NewIndex("a"), NewIndex("b"), false), b, root, obj{"a", 20, "b", 20})
	testOne(tb, NewAssign(NewIndex("a"), NewIndex("b"), true), b, root, obj{"a", 10, "b", 20})
}

func TestAssignArray(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", arr{nil, obj{"c", obj{"v", 10}}}, "v", 20})

	testOne(tb, NewAssign(NewIndex("a", 1, "c"), NewIndex("v"), false), b, root, obj{"a", arr{nil, obj{"c", 20}}, "v", 20})
	testOne(tb, NewAssign(NewIndex("a", 1, "c"), NewIndex("v"), true), b, root, obj{"a", arr{nil, obj{"c", 10}}, "v", 20})
}

func TestAssignArrayIter(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", arr{obj{"c", obj{"v", 5}}, obj{"c", obj{"v", 10}}}, "v", 20})

	testOne(tb, NewAssign(NewIndex("a", Iter{}, "c"), NewIndex("v"), false), b, root, obj{"a", arr{obj{"c", 20}, obj{"c", 20}}, "v", 20})
	testOne(tb, NewAssign(NewIndex("a", Iter{}, "c"), NewIndex("v"), true), b, root, obj{"a", arr{obj{"c", 5}, obj{"c", 10}}, "v", 20})
}

func TestAssignLongArray(tb *testing.T) {
	tb.Skip()

	b := NewBuffer(nil)
	root := b.appendVal(arr{
		obj{"a", arr{1, 1, 1, 1}},
		obj{"a", arr{2, 2, 2, 2, 2}},
		obj{"a", arr{3, 3, 3, 3}},
		//	obj{"a", arr{4, 4, 4, 4, 4, 4}},
		//	obj{"a", arr{5, 5, 5, 5, 5}},
	})

	exp := b.appendVal(arr{
		obj{"a", arr{0, 0, 0, 0}},
		obj{"a", arr{0, 0, 0, 0, 0}},
		obj{"a", arr{0, 0, 0, 0}},
		//	obj{"a", arr{0, 0, 0, 0, 0, 0}},
		//	obj{"a", arr{0, 0, 0, 0, 0}},
	})

	testOne(tb, NewAssign(NewIndex(Iter{}, "a", Iter{}), Off(Zero), false), b, root, Code(exp))
}
