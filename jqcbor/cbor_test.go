package jqcbor

import (
	"bytes"
	"encoding/hex"
	"testing"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

func TestDecodeEncode(tb *testing.T) {
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

	b := jq.NewBuffer()
	var d Decoder

	off, i, err := d.Decode(b, data, 0)
	assertNoError(tb, err)
	assertEqual(tb, len(data), i)

	f := jq.NewArray(jq.NewQuery(jq.Iter{}, jq.Iter{}, "c", jq.Iter{}))

	res, _, err := f.ApplyTo(b, off, false)
	assertNoError(tb, err)

	var e Encoder

	enc, err := e.Encode(nil, b, res)
	assertNoError(tb, err)

	if !bytes.Equal([]byte{byte(cbor.Array) | 6, 1, 2, 3, 4, 5, 6}, enc) {
		tb.Errorf("wanted array of 1 to 6, got % x", enc)
	}

	if tb.Failed() {
		tb.Logf("res %x -> %x\n%s", off, res, b.Dump())
	}
}

func TestFilter(tb *testing.T) {
	data := func() []byte {
		// {"a":"b","c":"d"}

		var e cbor.Encoder

		var b []byte

		b = e.AppendMap(b, 2)

		b = e.AppendString(b, "a")
		b = e.AppendString(b, "b")

		b = e.AppendString(b, "c")
		b = e.AppendString(b, "d")

		return b
	}()

	b := func() *jq.Buffer {
		b := jq.NewBuffer()

		b.B = b.Encoder.CBOR.AppendTag(b.B, cbor.String, len(data))
		b.B = append(b.B, data...)

		return b
	}()

	f := jq.NewPipe(
		NewDecoder(),
		NewEncoder(),
	//	NewEncoder(),
	)

	res, _, err := f.ApplyTo(b, 0, false)
	assertNoError(tb, err)

	s := b.Reader().Bytes(res)

	if !bytes.Equal([]byte(data), s) {
		tb.Errorf("not equal (% x) != (% x)", data, s)
	}

	if tb.Failed() {
		tb.Logf("hex\n%s", hex.Dump(b.B))
		tb.Logf("buffer\n%s", b.Dump())
	}
}

func TestLabeled(tb *testing.T) {
	data := func() []byte {
		// {"a":"b","c":"d"}

		var e cbor.Encoder

		var b []byte

		b = e.AppendMap(b, 2)

		b = e.AppendString(b, "a")
		b = e.AppendLabel(b, 10)
		b = e.AppendString(b, "b")

		b = e.AppendString(b, "c")
		b = e.AppendLabel(b, 100)
		b = e.AppendArray(b, 2)
		b = e.AppendString(b, "d")
		b = e.AppendString(b, "e")

		return b
	}()

	b := jq.NewBuffer()

	var d Decoder

	off, i, err := d.Decode(b, data, 0)
	assertNoError(tb, err)
	assertEqual(tb, len(data), i)

	var e Encoder

	enc, err := e.Encode(nil, b, off)
	assertNoError(tb, err)

	if !bytes.Equal(data, enc) {
		tb.Errorf("wanted % x, got % x\n%s", data, enc, b.Dump())
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
