package jq

import "testing"

func TestIndex(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(arr{"a", "b", "c", "d"})

	testOne(tb, Index(0), b, root, "a")
	testOne(tb, Index(1), b, root, "b")
	testOne(tb, Index(3), b, root, "d")
	testOne(tb, Index(-1), b, root, "d")
	testOne(tb, Index(-3), b, root, "b")
	testOne(tb, Index(-4), b, root, "a")
	testOne(tb, Index(-100), b, root, Null)
	testOne(tb, Index(100), b, root, Null)
}

func TestKey(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", 1, "b", 2, "c", 3, "d", 4})

	testOne(tb, Key("a"), b, root, 1)
	testOne(tb, Key("b"), b, root, 2)
	testOne(tb, Key("d"), b, root, 4)
	testOne(tb, Key("e"), b, root, Null)
}

func TestIndexPath(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(arr{"a", "b", "c", "d"})

	testOnePath(tb, Index(0), b, root, "a", NodePath{ps(root, 0)})
	testOnePath(tb, Index(1), b, root, "b", NodePath{ps(root, 1)})
	testOnePath(tb, Index(3), b, root, "d", NodePath{ps(root, 3)})
	testOnePath(tb, Index(-1), b, root, "d", NodePath{ps(root, 3)})
	testOnePath(tb, Index(-3), b, root, "b", NodePath{ps(root, 1)})
	testOnePath(tb, Index(-4), b, root, "a", NodePath{ps(root, 0)})
	testOnePath(tb, Index(-100), b, root, Null, NodePath{ps(root, -100)})
	testOnePath(tb, Index(100), b, root, Null, NodePath{ps(root, 100)})
}

func TestKeyPath(tb *testing.T) {
	b := NewBuffer()
	ra := b.appendVal("a")
	rb := b.appendVal("b")
	rd := b.appendVal("d")
	root := b.appendVal(obj{"a", 1, "b", 2, "c", 3, "d", 4})
	ekey := b.appendVal("e")

	testOnePath(tb, Key("a"), b, root, 1, NodePath{psk(root, 0, ra)})
	testOnePath(tb, Key("b"), b, root, 2, NodePath{psk(root, 1, rb)})
	testOnePath(tb, Key("d"), b, root, 4, NodePath{psk(root, 3, rd)})
	testOnePath(tb, Key("e"), b, root, Null, NodePath{psk(root, -1, ekey)})
}

func TestIndexLabeled(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(lab{lab: 100, val: arr{"a", "b", lab{lab: 10, val: "c"}, "d"}})

	testOne(tb, Index(0), b, root, "a")
	testOne(tb, Index(1), b, root, "b")
	testOne(tb, Index(-2), b, root, lab{lab: 10, val: "c"})
}

func TestKeyLabeled(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(lab{lab: 100, val: obj{"a", 1, "b", lab{lab: 10, val: 2}, "c", 3, "d", 4}})

	testOne(tb, Key("a"), b, root, 1)
	testOne(tb, Key("b"), b, root, lab{lab: 10, val: 2})
	testOne(tb, Key("d"), b, root, 4)
	testOne(tb, Key("e"), b, root, Null)
}

func TestIndexPathLabeled(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(lab{lab: 100, val: arr{"a", "b", lab{lab: 10, val: "c"}, "d"}})

	testOnePath(tb, Index(0), b, root, "a", NodePath{ps(root, 0)})
	testOnePath(tb, Index(1), b, root, "b", NodePath{ps(root, 1)})
	testOnePath(tb, Index(-2), b, root, lab{lab: 10, val: "c"}, NodePath{ps(root, 2)})
}

func TestKeyPathLabeled(tb *testing.T) {
	b := NewBuffer()
	ra := b.appendVal("a")
	rb := b.appendVal("b")
	rd := b.appendVal("d")
	root := b.appendVal(lab{lab: 200, val: obj{ra, 1, rb, lab{lab: 20, val: 2}, "c", 3, "d", 4}})
	ekey := b.appendVal("e")

	testOnePath(tb, Key("a"), b, root, 1, NodePath{psk(root, 0, ra)})
	testOnePath(tb, Key("b"), b, root, lab{lab: 20, val: 2}, NodePath{psk(root, 1, rb)})
	testOnePath(tb, Key("d"), b, root, 4, NodePath{psk(root, 3, rd)})
	testOnePath(tb, Key("e"), b, root, Null, NodePath{psk(root, -1, ekey)})
}

func ps(off Off, i int) NodePathSeg           { return NodePathSeg{Off: off, Index: i, Key: None} }
func psk(off Off, i int, key Off) NodePathSeg { return NodePathSeg{Off: off, Index: i, Key: key} }
