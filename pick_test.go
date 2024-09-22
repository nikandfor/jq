package jq

import "testing"

func TestPick(tb *testing.T) {
	b := NewBuffer(nil)
	r := b.appendVal(obj{"a", arr{obj{"a", 1}, obj{"a", 2, "b", 3}}})

	testOne(tb, NewPick(NewQuery("a", Iter{}, "a")), b, r, obj{"a", arr{obj{"a", 1}, obj{"a", 2}}})
}

func TestPickArray(tb *testing.T) {
	b := NewBuffer(nil)
	r := b.appendVal(arr{"a", "b", "c", "d", "e"})

	testOne(tb, NewPick(NewMulti(-5, 1, 4, 5)), b, r, arr{"a", "b", nil, nil, "e", nil})
}

func TestPickMap(tb *testing.T) {
	b := NewBuffer(nil)
	r := b.appendVal(obj{"a", "b", "c", "d", "e", "f"})

	testOne(tb, NewPick(NewMulti("e", "a")), b, r, obj{"e", "f", "a", "b"})
}

func TestPickMapAdd(tb *testing.T) {
	b := NewBuffer(nil)
	r := b.appendVal(obj{"a", "b", "c", "d", "e", "f"})

	testOne(tb, NewPick(NewMulti("c", "g", "g")), b, r, obj{"c", "d", "g", nil})
}

func TestPickIterAdd(tb *testing.T) {
	b := NewBuffer(nil)
	r := b.appendVal(obj{"a", arr{obj{"a", 1}, obj{"a", 2, "b", 3}}})

	testOne(tb, NewPick(NewQuery("a", Iter{}, "b")), b, r, obj{"a", arr{obj{"b", nil}, obj{"b", 3}}})
}

func TestPickIterAppend(tb *testing.T) {
	b := NewBuffer(nil)
	r := b.appendVal(obj{"a", arr{obj{"a", 1}, obj{"a", 2, "b", 3}}})

	testOne(tb, NewPick(NewQuery("a", NewMulti(Iter{}, 3), "b")), b, r, obj{"a", arr{obj{"b", nil}, obj{"b", 3}, nil, obj{"b", nil}}})
}

func TestPickCreate(tb *testing.T) {
	b := NewBuffer(nil)
	r := Null

	testOne(tb, NewPick(NewQuery(
		"a",
		3,
		"b",
	)), b, r, obj{
		"a",
		arr{
			nil, nil, nil,
			obj{"b", nil},
		},
	})
}
