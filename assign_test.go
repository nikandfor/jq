package jq

import "testing"

func TestAssignObjectAbs(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", obj{"b", 10}, "b", 20})

	testOne(tb, NewAssign(Key("a"), Key("b"), false), b, off, obj{"a", 20, "b", 20})
}

func TestAssignObjectRel(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", obj{"b", 10}, "b", 20})

	testOne(tb, NewAssign(Key("a"), Key("b"), true), b, off, obj{"a", 10, "b", 20})
}

func TestAssignArrayAbs(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", arr{nil, obj{"c", obj{"v", 10}}}, "v", 20})

	testOne(tb, NewAssign(NewQuery("a", 1, "c"), Key("v"), false), b, off, obj{"a", arr{nil, obj{"c", 20}}, "v", 20})
}

func TestAssignArrayRel(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", arr{nil, obj{"c", obj{"v", 10}}}, "v", 20})

	testOne(tb, NewAssign(NewQuery("a", 1, "c"), Key("v"), true), b, off, obj{"a", arr{nil, obj{"c", 10}}, "v", 20})
}

func TestAssignArrayIterAbs(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", arr{obj{"c", obj{"v", 5}}, obj{"c", obj{"v", 10}}}, "v", 20})

	testOne(tb, NewAssign(NewQuery("a", Iter{}, "c"), Key("v"), false), b, off, obj{"a", arr{obj{"c", 20}, obj{"c", 20}}, "v", 20})
}

func TestAssignArrayIterRel(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", arr{obj{"c", obj{"v", 5}}, obj{"c", obj{"v", 10}}}, "v", 20})

	testOne(tb, NewAssign(NewQuery("a", Iter{}, "c"), Key("v"), true), b, off, obj{"a", arr{obj{"c", 5}, obj{"c", 10}}, "v", 20})
}

func TestAssignLongArrayAbs(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(arr{
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

	testOne(tb, NewAssign(NewQuery(Iter{}, "a", Iter{}), Off(Zero), false), b, off, Code(exp))
}

func TestAssignComma(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", 1, "b", 2})
	exp := b.appendVal(obj{"a", 10, "b", 10})

	testOne(tb, NewAssign(NewComma(Key("a"), Key("b")), NewLiteral(10), false), b, off, code(exp))
}

func TestAssignCommaPipe(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", 1, "b", obj{"c", 2, "d", obj{"e", 3}}, "q", 4})
	exp := b.appendVal(obj{"a", 10, "b", obj{"c", 10, "d", obj{"e", 10}}, "q", 10})

	testOne(tb,
		NewAssign(
			NewComma(
				Key("a"),
				NewPipe(
					Key("b"),
					NewComma(
						Key("c"),
						NewPipe(Key("d"), Key("e")),
					),
				),
				Key("q"),
			),
			NewLiteral(10), false,
		),
		b, off, code(exp),
	)
}

func TestAssignSelect(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(arr{obj{"a", false, "v", 1}, obj{"a", true, "v", 2}, obj{"a", 1, "v", 3}, obj{"a", nil, "v", 4}})
	exp := b.appendVal(arr{obj{"a", false, "v", 1}, obj{"a", true, "v", 10}, obj{"a", 1, "v", 10}, obj{"a", nil, "v", 4}})

	testOne(tb, NewAssign(NewPipe(NewIter(), NewSelect(Key("a")), Key("v")), NewLiteral(10), false), b, off, code(exp))
}

func TestAssignMulti(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", arr{1, 2}, "b", arr{"q", "w"}})

	testIter(tb, NewAssign(NewQuery("a", Iter{}), NewQuery("b", Iter{}), false), b, off, []any{
		obj{"a", arr{"q", "q"}, "b", arr{"q", "w"}},
		obj{"a", arr{"w", "w"}, "b", arr{"q", "w"}},
	})
}

func TestAssignNoneAbs(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", 1, "b", 2})

	testIter(tb, NewAssign(Key("b"), Off(None), false), b, off, nil)
}

func TestAssignNoneRel(tb *testing.T) {
	b := NewBuffer(nil)
	off := b.appendVal(obj{"a", 1, "b", 2})
	exp := b.appendVal(obj{"a", 1})

	testOne(tb, NewAssign(Key("b"), Off(None), true), b, off, code(exp))
}
