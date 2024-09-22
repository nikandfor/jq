package jq

import "testing"

func TestPipe(tb *testing.T) {
	d, root := appendValBuf(nil, 0, obj{"a", obj{"b", obj{"c", "d"}}})
	b := NewBuffer(d)

	testOne(tb, NewPipe(Key("a"), Key("b"), Key("c")), b, root, "d")

	if tb.Failed() {
		tb.Logf("buffer\n%s", Dump(b))
	}

	d, root = appendValBuf(d, 0, "a")
	b.Reset(d)

	testIter(tb, NewPipe(
		NewComma(Dot{}, Dot{}),
		NewComma(Dot{}, Dot{}),
	), b, root, []any{"a", "a", "a", "a"})

	d, root = appendValBuf(d, 0, arr{arr{arr{"a", "b"}, arr{"c", "d"}}})
	b.Reset(d)

	testIter(tb, NewPipe(
		&Iter{}, &Iter{}, &Iter{},
	), b, root, []any{"a", "b", "c", "d"})
}

func TestPipePathABC(tb *testing.T) {
	b := NewBuffer(nil)
	r2 := b.appendVal(obj{"c", "d"})
	r1 := b.appendVal(obj{"b", r2})
	r0 := b.appendVal(obj{"a", r1})

	testOnePath(tb, NewPipe(Key("a"), Key("b"), Key("c")), b, r0, "d", NodePath{ps(r0, 0), ps(r1, 0), ps(r2, 0)})

	if tb.Failed() {
		tb.Logf("buffer\n%s", Dump(b))
	}
}

func TestPipePathAAAA(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal("a")

	testIterPath(tb, NewPipe(
		NewComma(Dot{}, Dot{}),
		NewComma(Dot{}, Dot{}),
	), b, root, []any{"a", "a", "a", "a"}, []NodePath{nil, nil, nil, nil})

	if tb.Failed() {
		tb.Logf("buffer\n%s", Dump(b))
	}
}

func TestPipePathABCD(tb *testing.T) {
	b := NewBuffer(nil)
	r10 := b.appendVal(arr{"a", "b"})
	r11 := b.appendVal(arr{"c", "d"})
	r1 := b.appendVal(arr{r10, r11})
	r0 := b.appendVal(arr{r1})

	testIterPath(tb, NewPipe(
		&Iter{}, &Iter{}, &Iter{},
	), b, r0, []any{"a", "b", "c", "d"},
		[]NodePath{
			{ps(r0, 0), ps(r1, 0), ps(r10, 0)},
			{ps(r0, 0), ps(r1, 0), ps(r10, 1)},
			{ps(r0, 0), ps(r1, 1), ps(r11, 0)},
			{ps(r0, 0), ps(r1, 1), ps(r11, 1)},
		})

	if tb.Failed() {
		tb.Logf("buffer\n%s", Dump(b))
	}
}
