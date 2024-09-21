package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestFirstLastNth(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(arr{0, 1, 2, 3})

	testOne(tb, NewArray(NewComma(NewFirst(), NewLast(), NewNth(2))), b, b.appendVal(arr{}), arr{nil, nil, nil})
	testOne(tb, NewArray(NewComma(NewFirst(), NewLast(), NewNth(2))), b, root, arr{0, 3, 2})

	expr := NewComma(Literal{0}, Literal{1}, Empty{}, Literal{2}, Literal{cbor.Simple | cbor.Null}, False)
	testOne(tb, NewArray(NewComma(NewFirstOf(expr), NewLastOf(expr), NewNthOf(expr, 2), NewNthOf(expr, 3))), b, root, arr{0, false, 2, nil})

	testOne(tb, NewNthOf(expr, -10), b, None, None)
	testOne(tb, NewNthOf(NewComma(Literal{1}, Literal{2}, Literal{3}), 5), b, None, None)
}
