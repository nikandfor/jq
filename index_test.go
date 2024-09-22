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

	testOnePath(tb, Index(0), b, root, "a", NodePath{ps(root, 0)})
	testOnePath(tb, Index(1), b, root, "b", NodePath{ps(root, 1)})
	testOnePath(tb, Index(3), b, root, "d", NodePath{ps(root, 3)})
	testOnePath(tb, Index(-1), b, root, "d", NodePath{ps(root, 3)})
	testOnePath(tb, Index(-3), b, root, "b", NodePath{ps(root, 1)})
	testOnePath(tb, Index(-4), b, root, "a", NodePath{ps(root, 0)})
	testOnePath(tb, Index(-100), b, root, code(Null), NodePath{ps(root, -100)})
	testOnePath(tb, Index(100), b, root, code(Null), NodePath{ps(root, 100)})
}

func TestKeyPath(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", 1, "b", 2, "c", 3, "d", 4})
	ekey := b.appendVal("e")

	testOnePath(tb, Key("a"), b, root, 1, NodePath{ps(root, 0)})
	testOnePath(tb, Key("b"), b, root, 2, NodePath{ps(root, 1)})
	testOnePath(tb, Key("d"), b, root, 4, NodePath{ps(root, 3)})
	testOnePath(tb, Key("e"), b, root, code(Null), NodePath{psf(root, -1, ekey)})
}

func ps(off Off, i int) NodePathSeg           { return NodePathSeg{Off: off, Index: i, Key: None} }
func psf(off Off, i int, key Off) NodePathSeg { return NodePathSeg{Off: off, Index: i, Key: key} }
