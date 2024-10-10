package jq

import "testing"

func TestKeys(tb *testing.T) {
	b := NewBuffer()
	r := b.appendVal(obj{"c", 3, "d", 4, "b", 2, "a", 1})

	testOne(tb, NewKeys(), b, r, arr{"a", "b", "c", "d"})
}
