package jq

import "testing"

func TestIter(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", 1, "b", 2, "c", "d"})

	testIter(tb, NewIter(), b, root, []any{1, 2, "d"})
}

func TestIterPath(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", 1, "b", 2, "c", "d"})

	testIterPath(tb, NewIter(), b, root, []any{1, 2, "d"}, []NodePath{{ps(root, 0)}, {ps(root, 1)}, {ps(root, 2)}})
}
