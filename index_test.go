package jq

import "testing"

func TestIndex(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(arr{"a", "b", "c", "d"})

	testOne(tb, Index(0), b, root, "a")
	testOne(tb, Index(1), b, root, "b")
	testOne(tb, Index(3), b, root, "d")
	testOne(tb, Index(-1), b, root, "d")
	testOne(tb, Index(-3), b, root, "b")
	testOne(tb, Index(-4), b, root, "a")
	testOne(tb, Index(-100), b, root, code(Null))
	testOne(tb, Index(100), b, root, code(Null))
}

func TestKey(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", 1, "b", 2, "c", 3, "d", 4})

	testOne(tb, Key("a"), b, root, 1)
	testOne(tb, Key("b"), b, root, 2)
	testOne(tb, Key("d"), b, root, 4)
	testOne(tb, Key("e"), b, root, code(Null))
}

func TestIndexPath(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(arr{"a", "b", "c", "d"})

	testOnePath(tb, Index(0), b, root, "a", Path{ps(root, 0)})
	testOnePath(tb, Index(1), b, root, "b", Path{ps(root, 1)})
	testOnePath(tb, Index(3), b, root, "d", Path{ps(root, 3)})
	testOnePath(tb, Index(-1), b, root, "d", Path{ps(root, 3)})
	testOnePath(tb, Index(-3), b, root, "b", Path{ps(root, 1)})
	testOnePath(tb, Index(-4), b, root, "a", Path{ps(root, 0)})
	testOnePath(tb, Index(-100), b, root, code(Null), Path{ps(root, -1)})
	testOnePath(tb, Index(100), b, root, code(Null), Path{ps(root, -1)})
}

func TestKeyPath(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", 1, "b", 2, "c", 3, "d", 4})

	testOnePath(tb, Key("a"), b, root, 1, Path{ps(root, 0)})
	testOnePath(tb, Key("b"), b, root, 2, Path{ps(root, 1)})
	testOnePath(tb, Key("d"), b, root, 4, Path{ps(root, 3)})
	testOnePath(tb, Key("e"), b, root, code(Null), Path{ps(root, -1)})
}

func ps(off Off, i int) PathSeg { return PathSeg{Off: off, Index: i} }
