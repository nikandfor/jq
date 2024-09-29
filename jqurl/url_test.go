package jqurl

import (
	"bytes"
	"testing"

	"nikand.dev/go/jq"
)

func TestEncoder(tb *testing.T) {
	b := jq.NewBuffer()
	v := b.AppendValue(jq.Obj{"a", "b", "c", 4, "d", false, "e", 1.1, "f", nil})

	e := NewEncoder()

	res, err := e.Encode(nil, b, v)
	assertNoError(tb, err)
	assertBytes(tb, []byte(`a=b&c=4&d=false&e=1.1&f`), res)
}

func TestEncoderIter(tb *testing.T) {
	b := jq.NewBuffer()
	v := b.AppendValue(jq.Obj{"a", jq.Arr{"b", 4, false, 1.1, nil}})

	e := NewEncoder()

	res, err := e.Encode(nil, b, v)
	assertNoError(tb, err)
	assertBytes(tb, []byte(`a=b&a=4&a=false&a=1.1&a`), res)
}

func TestDecoderEncoder(tb *testing.T) {
	b := jq.NewBuffer()

	d := NewDecoder()
	e := NewEncoder()

	for _, x := range []string{
		"",
		"param",
		"param=val",
		"a=b&c=d&e",
		"a=b&a=c&a=d",
	} {
		b.Reset()

		off, i, err := d.Decode(b, []byte(x), 0)
		assertNoError(tb, err)
		assertEqual(tb, len(x), i)

		y, err := e.Encode(nil, b, off)
		assertNoError(tb, err)

		assertBytes(tb, []byte(x), y)

		tb.Logf("dump  %v  (%s)\n%s", off, x, b.Dump())

		if tb.Failed() {
			tb.Logf("dump  %v  (%s)\n%s", off, x, b.Dump())
			break
		}
	}
}

func assertNoError(tb testing.TB, err error) {
	tb.Helper()

	if err != nil {
		tb.Errorf("unexpected error: %v", err)
	}
}

func assertBytes(tb testing.TB, x, y []byte) {
	tb.Helper()

	if !bytes.Equal(x, y) {
		tb.Errorf("not equal (%s) != (%s)", x, y)
	}
}

func assertEqual(tb testing.TB, x, y any) {
	tb.Helper()

	if x != y {
		tb.Errorf("expected to be equal: %v and %v", x, y)
	}
}
