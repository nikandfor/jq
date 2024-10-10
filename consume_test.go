package jq

import "testing"

func TestConsume(tb *testing.T) {
	b := NewBuffer()
	r := b.appendVal(obj{"a", 1, "b", "str"})

	var av Int
	var bv Bytes

	testOne(tb, NewMulti(NewQuery("a", &av), NewQuery("b", &bv)), b, r, None)

	if av != 1 || string(bv) != "str" {
		tb.Errorf("wrong %v %v", av, bv)
	}
}
