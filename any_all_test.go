package jq

import "testing"

func TestAnyAll(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		arr{},
		arr{false, nil},
		arr{false, nil, 1},
		arr{1, 2, 3},
		obj{},
		obj{"a", false, "b", nil},
		obj{"a", false, "b", nil, "c", 1},
		obj{"a", 1},
	})

	testIter(tb, NewPipe(NewIter(), NewAny(nil, nil)), b, root, []any{
		false, false, true, true,
		false, false, true, true,
	})

	testIter(tb, NewPipe(NewIter(), NewAny(NewIter(), NewDot())), b, root, []any{
		false, false, true, true,
		false, false, true, true,
	})

	testIter(tb, NewPipe(NewIter(), NewAll(nil, nil)), b, root, []any{
		true, false, false, true,
		true, false, false, true,
	})

	testIter(tb, NewPipe(NewIter(), NewAll(NewIter(), NewDot())), b, root, []any{
		true, false, false, true,
		true, false, false, true,
	})

	root = b.appendVal(arr{
		arr{1, 2, 3},
		arr{2},
		arr{1},
	})

	testIter(tb, NewPipe(NewIter(), NewAny(NewIter(), NewComma(NewEqual(NewDot(), NewLiteral(-1)), NewEqual(NewDot(), NewLiteral(2))))), b, root, []any{
		true, true, false,
	})

	root = b.appendVal(arr{
		arr{1, 2, 3},
		arr{2},
		arr{1},
	})

	testIter(tb, NewPipe(NewIter(), NewAll(NewIter(), NewComma(NewEqual(NewDot(), NewLiteral(-1)), NewEqual(NewDot(), NewLiteral(2))))), b, root, []any{
		false, false, false,
	})

	testIter(tb, NewPipe(NewIter(), NewAll(NewIter(), NewComma(NewEqual(NewDot(), NewLiteral(2)), NewEqual(NewDot(), NewLiteral(2))))), b, root, []any{
		false, true, false,
	})
}
