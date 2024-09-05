package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestIsSimpleAppendVal(tb *testing.T) {
	b := &Buffer{}

	for j, x := range []int{None, Null, True, False, Zero, One} {
		off := b.appendVal(code(x))

		assertEqualVal(tb, b, x, off, "%x", x)

		assertTrue(tb, b.Reader().IsSimple(x, x), "%x", x)
		assertTrue(tb, b.Reader().IsSimple(off, x), "%x", x)

		assertTrue(tb, !b.Reader().IsSimple(x, x+1), "%x", x)
		assertTrue(tb, !b.Reader().IsSimple(off, x+1), "%x", x)

		assertTrue(tb, b.Reader().IsSimple(off, Zero, One, True, False, Null, None), "j: %x  cbor: %x  code: all", j, x)

	}

	tb.Logf("buffer\n%s", Dump(b))
}

func TestIsSimpleCBOREncoder(tb *testing.T) {
	var r []byte
	b := NewBuffer(r)

	for j, tc := range []struct {
		CBOR byte
		Code int
	}{
		{CBOR: cbor.Int | 0, Code: Zero},
		{CBOR: cbor.Int | 1, Code: One},
		{CBOR: cbor.Simple | cbor.None, Code: None},
		{CBOR: cbor.Simple | cbor.False, Code: False},
		{CBOR: cbor.Simple | cbor.True, Code: True},
		{CBOR: cbor.Simple | cbor.Null, Code: Null},
	} {
		off := len(r)
		r = append(r, tc.CBOR)
		b.Reset(r)

		assertTrue(tb, b.Reader().IsSimple(off, tc.Code), "j: %x  cbor: %x  code: %x", j, tc.CBOR, tc.Code)
		assertTrue(tb, b.Reader().IsSimple(off, Zero, One, True, False, Null, None), "j: %x  cbor: %x  code: all", j, tc.CBOR)
	}

	tb.Logf("buffer\n%s", Dump(b))
}
