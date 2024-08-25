package jqcbor

import (
	"bytes"
	"testing"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

func TestCBOR(tb *testing.T) {
	data := func() []byte {
		// `{"a":[{"q":"w","c":[1,2,3]},{"c":[4],"d":44}],"b":[{"c":[]},{"c":[5,6]}]}`

		var e cbor.Encoder

		var b []byte

		b = e.AppendMap(b, -1)
		b = e.AppendString(b, "a")

		{
			b = e.AppendArray(b, 2)

			b = e.AppendMap(b, 2)

			b = e.AppendString(b, "q")
			b = e.AppendString(b, "w")

			b = e.AppendString(b, "c")
			b = e.AppendArray(b, 3)
			b = append(b, 1, 2, 3)

			b = e.AppendMap(b, 2)

			b = e.AppendString(b, "c")
			b = e.AppendArray(b, 1)
			b = e.AppendInt(b, 4)

			b = e.AppendString(b, "d")
			b = e.AppendInt(b, 44)
		}

		b = e.AppendString(b, "b")

		{
			b = e.AppendArray(b, 2)

			b = e.AppendMap(b, 1)
			b = e.AppendString(b, "c")
			b = e.AppendArray(b, 0)

			b = e.AppendMap(b, 1)
			b = e.AppendString(b, "c")
			b = e.AppendArray(b, 2)
			b = append(b, 5, 6)
		}

		b = e.AppendBreak(b)

		return b
	}()

	var d Decoder

	w, off, i, err := d.Decode(nil, data, 0)
	assertNoError(tb, err)
	assertEqual(tb, len(data), i)

	b := jq.NewBuffer(w)

	f := jq.NewArray(jq.NewIndex(jq.Iter{}, jq.Iter{}, "c", jq.Iter{}))

	res, _, err := f.ApplyTo(b, off, false)
	assertNoError(tb, err)

	var e Encoder

	r0, r1 := b.Unwrap()
	enc := e.Encode(nil, r0, r1, res)

	if !bytes.Equal([]byte{cbor.Array | 6, 1, 2, 3, 4, 5, 6}, enc) {
		tb.Errorf("wanted array of 1 to 6, got % x", enc)
	}

	if tb.Failed() {
		tb.Logf("res %x -> %x\n%s", off, res, jq.DumpBuffer(b))
	}
}

func assertNoError(tb testing.TB, err error) {
	tb.Helper()

	if err != nil {
		tb.Errorf("unexpected error: %v", err)
	}
}

func assertEqual(tb testing.TB, x, y any) {
	tb.Helper()

	if x != y {
		tb.Errorf("expected to be equal: %v and %v", x, y)
	}
}
