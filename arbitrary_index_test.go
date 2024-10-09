package jq

import "testing"

func TestArbitraryIndex(tb *testing.T) {
	b := NewBuffer()
	r0 := b.appendVal(obj{
		"arr",
		arr{"a", "b", "c", "d"},
		"idx",
		arr{0, 1, 3, -1, -3, -4, -100, 100},
	})

	testIter(tb, NewIndex(NewQuery("arr"), NewQuery("idx", Iter{})), b, r0, []any{
		"a", "b", "d", "d", "b", "a", Null, Null,
	})
}

func TestArbitraryKey(tb *testing.T) {
	b := NewBuffer()
	r1 := b.appendVal(obj{
		"a",
		arr{
			obj{"a", 1, "b", 2},
			obj{"a", 3, "b", 4},
		},
		"b",
		arr{"a", "b"},
	})

	testIter(tb, NewIndex(NewQuery("a", Iter{}), NewQuery("b", Iter{})), b, r1, []any{
		1, 2, 3, 4,
	})
}

func TestArbitraryIndexPath(tb *testing.T) {
	b := NewBuffer()

	arrkey := b.appendVal("arr")
	a0 := b.appendVal("a")
	b0 := b.appendVal("b")
	c0 := b.appendVal("c")
	d0 := b.appendVal("d")
	arr0 := b.appendVal(arr{a0, b0, c0, d0})

	r0 := b.appendVal(obj{
		arrkey,
		arr0,
		"idx",
		arr{0, 1, 3, -1, -3, -4, -100, 100},
	})

	testIterPath(tb, NewIndex(NewQuery("arr"), NewQuery("idx", Iter{})), b, r0, []any{
		"a", "b", "d", "d", "b", "a", Null, Null,
	}, []NodePath{
		{psk(r0, 0, arrkey), ps(arr0, 0)},
		{psk(r0, 0, arrkey), ps(arr0, 1)},
		{psk(r0, 0, arrkey), ps(arr0, 3)},
		{psk(r0, 0, arrkey), ps(arr0, 3)},
		{psk(r0, 0, arrkey), ps(arr0, 1)},
		{psk(r0, 0, arrkey), ps(arr0, 0)},
		{psk(r0, 0, arrkey), ps(arr0, -100)},
		{psk(r0, 0, arrkey), ps(arr0, 100)},
	})

	if tb.Failed() {
		tb.Logf("dump %v\n%s", r0, b.Dump())
	}
}

func TestArbitraryKeyPath(tb *testing.T) {
	b := NewBuffer()

	akey := b.appendVal("a")
	bkey := b.appendVal("b")
	ckey := b.appendVal("c")

	obj0 := b.appendVal(obj{akey, 1, bkey, 2})
	obj1 := b.appendVal(obj{akey, 3, bkey, 4})

	arr0 := b.appendVal(arr{obj0, obj1})

	r0 := b.appendVal(obj{
		bkey,
		arr{akey, bkey, ckey},
		akey,
		arr0,
	})

	testIterPath(tb, NewIndex(NewQuery("a", Iter{}), NewQuery("b", Iter{})), b, r0, []any{
		1, 2, Null,
		3, 4, Null,
	}, []NodePath{
		{psk(r0, 1, akey), ps(arr0, 0), psk(obj0, 0, akey)},
		{psk(r0, 1, akey), ps(arr0, 0), psk(obj0, 1, bkey)},
		{psk(r0, 1, akey), ps(arr0, 0), psk(obj0, -1, ckey)},

		{psk(r0, 1, akey), ps(arr0, 1), psk(obj1, 0, akey)},
		{psk(r0, 1, akey), ps(arr0, 1), psk(obj1, 1, bkey)},
		{psk(r0, 1, akey), ps(arr0, 1), psk(obj1, -1, ckey)},
	})

	if tb.Failed() {
		tb.Logf("dump %v\n%s", r0, b.Dump())
	}
}
