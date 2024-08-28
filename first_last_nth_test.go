package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestFirstLastNth(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(arr{0, 1, 2, 3})

	testOne(tb, NewArray(NewComma(First{}, Last{}, Nth{N: 2})), b, b.appendVal(arr{}), arr{nil, nil, nil})
	testOne(tb, NewArray(NewComma(First{}, Last{}, Nth{N: 2})), b, root, arr{0, 3, 2})

	expr := NewComma(Literal{0}, Literal{1}, Empty{}, Literal{2}, Literal{cbor.Simple | cbor.Null}, Off(False))
	testOne(tb, NewArray(NewComma(NewFirst(expr), NewLast(expr), NewNth(2, expr), NewNth(3, expr))), b, root, arr{0, false, 2, nil})

	testOne(tb, NewNth(5, nil), b, b.appendVal(arr{1, 2}), nil)
	testOne(tb, NewNth(5, NewComma(Literal{1}, Literal{2}, Literal{3})), b, None, Off(None))
}
