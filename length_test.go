package jq

import "testing"

func TestLength(tb *testing.T) {
	b := NewBuffer(nil)

	root := b.appendVal(arr{
		arr{},
		arr{1, 2, 3},
		obj{},
		obj{1, 1, 2, 2, 3, 3},
		"",
		"qwerty",
		[]byte{},
		[]byte("qwerty"),
		5, -5,
		nil,
		1.0,
		-1.0,
	})

	testIter(tb, NewPipe(&Iter{}, Length{}), b, root, []any{
		0, 3, 0, 3, 0, 6, 0, 6, 5, 5, 0, 1.0, 1.0,
	})
}
