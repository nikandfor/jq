package jq

import "testing"

func TestPipe(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal(obj{"a", obj{"b", obj{"c", "d"}}})

	testOne(tb, NewPipe(Key("a"), Key("b"), Key("c")), b, root, "d")

	if tb.Failed() {
		tb.Logf("buffer\n%s", b.Dump())
	}

	b.Reset()
	root = b.appendVal("a")

	testIter(tb, NewPipe(
		NewComma(Dot{}, Dot{}),
		NewComma(Dot{}, Dot{}),
	), b, root, []any{"a", "a", "a", "a"})

	b.Reset()
	root = b.appendVal(arr{arr{arr{"a", "b"}, arr{"c", "d"}}})

	testIter(tb, NewPipe(
		&Iter{}, &Iter{}, &Iter{},
	), b, root, []any{"a", "b", "c", "d"})
}

func TestPipePathABC(tb *testing.T) {
	b := NewBuffer()

	ka := b.appendVal("a")
	kb := b.appendVal("b")
	kc := b.appendVal("c")

	r2 := b.appendVal(obj{kc, "d"})
	r1 := b.appendVal(obj{kb, r2})
	r0 := b.appendVal(obj{ka, r1})

	testOnePath(tb, NewPipe(Key("a"), Key("b"), Key("c")), b, r0, "d",
		NodePath{psk(r0, 0, ka), psk(r1, 0, kb), psk(r2, 0, kc)},
	)

	if tb.Failed() {
		tb.Logf("buffer\n%s", b.Dump())
	}
}

func TestPipePathAAAA(tb *testing.T) {
	b := NewBuffer()
	root := b.appendVal("a")

	testIterPath(tb, NewPipe(
		NewComma(Dot{}, Dot{}),
		NewComma(Dot{}, Dot{}),
	), b, root, []any{"a", "a", "a", "a"}, []NodePath{nil, nil, nil, nil})

	if tb.Failed() {
		tb.Logf("buffer\n%s", b.Dump())
	}
}

func TestPipePathABCD(tb *testing.T) {
	b := NewBuffer()
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
		tb.Logf("buffer\n%s", b.Dump())
	}
}
