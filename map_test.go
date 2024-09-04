package jq

import "testing"

func TestMapArr(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(arr{"a", 1})

	testOne(tb, NewMap(NewComma(Dot{}, Dot{})), b, root, arr{"a", "a", 1, 1})
	testOne(tb, NewMapValues(NewComma(Dot{}, Dot{})), b, root, arr{"a", 1})

	// tb.Logf("buffer\n%s", DumpBuffer(b))
}

func TestMapObj(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", "q", "b", 2})

	testOne(tb, NewMap(NewComma(Dot{}, Dot{})), b, root, arr{"q", "q", 2, 2})
	testOne(tb, NewMapValues(NewComma(Dot{}, Dot{})), b, root, obj{"a", "q", "b", 2})
}

func TestMapObjObj(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", obj{"x", "y", "y", "z"}, "b", obj{"x", "a", "z", arr{"b", "c"}}})

	testOne(tb, NewMap(NewComma(
		NewQuery("x"),
		NewQuery("y"),
		NewQuery("z"),
	)), b, root, arr{"y", "z", nil, "a", nil, arr{"b", "c"}})

	testOne(tb, NewMapValues(NewComma(
		NewQuery("x"),
		NewQuery("y"),
		NewQuery("z"),
	)), b, root, obj{"a", "y", "b", "a"})
}
