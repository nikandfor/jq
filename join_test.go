package jq

import "testing"

func TestJoin(tb *testing.T) {
	b := NewBuffer()
	r := b.appendVal(arr{"a", "b", "c"})
	sep := b.appendVal(", ")
	multi := b.appendVal(arr{":", "+"})

	testOne(tb, NewJoin(nil), b, r, "abc")
	testOne(tb, NewJoin(EmptyString), b, r, "abc")
	testOne(tb, NewJoin(sep), b, r, "a, b, c")
	testIter(tb, NewJoin(NewIterOf(multi)), b, r, []any{"a:b:c", "a+b+c"})

	o := b.appendVal(arr{obj{"a", "first", "b", "second"}, obj{"a", "12", "b", "34"}})

	testIter(tb, NewJoinOf(
		NewPipe(NewIter(), NewArray(NewComma(Key("a"), Key("b")))),
		NewIterOf(multi),
	), b, o, []any{
		"first:second",
		"first+second",
		"12:34",
		"12+34",
	})
}
