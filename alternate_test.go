package jq

import "testing"

func TestAlternate(tb *testing.T) {
	b := &Buffer{}
	root := b.appendVal(obj{
		"a",
		arr{},
		"b",
		arr{1, nil, false, 2},
		"c",
		arr{nil, false},
		"d", 5,
	})

	testIter(tb, NewAlternate(NewQuery("a", Iter{}), NewQuery("b", Iter{})), b, root, []any{1, nil, false, 2})
	testIter(tb, NewAlternate(NewQuery("b", Iter{}), NewQuery("d")), b, root, []any{1, 2})
	testIter(tb, NewAlternate(NewQuery("c", Iter{}), NewQuery("d")), b, root, []any{5})
}

func TestAlternatePath(tb *testing.T) {
	b := &Buffer{}

	ka := b.appendVal("a")
	kb := b.appendVal("b")
	kc := b.appendVal("c")
	kd := b.appendVal("d")

	ra := b.appendVal(arr{})
	rb := b.appendVal(arr{1, nil, false, 2})
	rc := b.appendVal(arr{nil, false})
	rd := b.appendVal(5)

	r0 := b.appendVal(obj{
		ka, ra,
		kb, rb,
		kc, rc,
		kd, rd,
	})

	testIterPath(tb, NewAlternate(NewQuery("a", Iter{}), NewQuery("b", Iter{})), b, r0,
		[]any{1, nil, false, 2},
		[]NodePath{
			{psk(r0, 1, kb), ps(rb, 0)},
			{psk(r0, 1, kb), ps(rb, 1)},
			{psk(r0, 1, kb), ps(rb, 2)},
			{psk(r0, 1, kb), ps(rb, 3)},
		})

	testIterPath(tb, NewAlternate(NewQuery("b", Iter{}), NewQuery("d")), b, r0,
		[]any{1, 2},
		[]NodePath{
			{psk(r0, 1, kb), ps(rb, 0)},
			{psk(r0, 1, kb), ps(rb, 3)},
		})

	testIterPath(tb, NewAlternate(NewQuery("c", Iter{}), NewQuery("d")), b, r0, []any{5}, []NodePath{{psk(r0, 3, kd)}})

	if tb.Failed() {
		tb.Logf("buffer:\n%s", Dump(b))
	}
}
