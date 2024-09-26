package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestFirstLastNth(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(arr{0, 1, 2, 3})

	testOne(tb, NewArray(NewComma(NewFirst(), NewLast(), NewNth(2))), b, b.appendVal(arr{}), arr{nil, nil, nil})
	testOne(tb, NewArray(NewComma(NewFirst(), NewLast(), NewNth(2))), b, root, arr{0, 3, 2})

	expr := NewComma(NewLiteral(0), NewLiteral(1), Empty{}, NewLiteral(2), Literal{[]byte{byte(cbor.Simple | cbor.Null)}}, False)
	testOne(tb, NewArray(NewComma(NewFirstOf(expr), NewLastOf(expr), NewNthOf(expr, 2), NewNthOf(expr, 3))), b, root, arr{0, false, 2, nil})

	testOne(tb, NewNthOf(expr, -10), b, None, None)
	testOne(tb, NewNthOf(NewComma(NewLiteral(1), NewLiteral(2), NewLiteral(3)), 5), b, None, None)
}

func TestLimit(tb *testing.T) {
	b := NewBuffer()
	arr := b.appendVal(arr{0, 1, 2, 3, 4})

	testIter(tb, NewLimit(NewIter(), 3), b, arr, []any{0, 1, 2})
	testIter(tb, NewLimit(NewComma(Index(-1), Index(-2), Index(-3), Key("fail")), 3), b, arr, []any{4, 3, 2})
}

func TestIsEmpty(tb *testing.T) {
	b := NewBuffer()
	arr1 := b.appendVal(arr{0, 1, 2, 3, 4})
	arr2 := b.appendVal(arr{})

	testOne(tb, NewIsEmpty(NewIter()), b, arr1, False)
	testOne(tb, NewIsEmpty(NewIter()), b, arr2, True)
	testOne(tb, NewIsEmpty(Empty{}), b, arr2, True)
}
