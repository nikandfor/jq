package jqjson

import (
	"bytes"
	"encoding/hex"
	"testing"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

func TestDecodeEncode(tb *testing.T) {
	data := []byte(`{"a":[{"q":"w","c":[1,2,3]},{"c":[4],"d":44}],"b":[{"c":[]},{"c":[5.,6.]}]}`)
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

	if !bytes.Equal([]byte(`[1,2,3,4,5,6]`), enc) {
		tb.Errorf("wanted array of 1 to 6, got %s", enc)
	}

	if tb.Failed() {
		tb.Logf("res %x -> %x\n%s", off, res, b.Dump())
	}
}

func TestFilter(tb *testing.T) {
	data := `{"a":"b","c":"d"}`

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
		tb.Errorf("not equal (%s) != (%s)", data, s)
	}

	if tb.Failed() {
		tb.Logf("hex\n%s", hex.Dump(b.B))
		tb.Logf("buffer\n%s", b.Dump())
	}
}

func TestMergeString(tb *testing.T) {
	b, root := func() (*jq.Buffer, Off) {
		b := jq.NewBuffer()

		b.B = b.Encoder.CBOR.AppendTag(b.B, cbor.String, -1)

		b.B = b.Encoder.CBOR.AppendString(b.B, "one_")

		b.B = b.Encoder.CBOR.AppendTag(b.B, cbor.String, -1)
		b.B = b.Encoder.CBOR.AppendString(b.B, "two_")
		b.B = b.Encoder.CBOR.AppendString(b.B, "three")
		b.B = b.Encoder.CBOR.AppendBreak(b.B)

		b.B = b.Encoder.CBOR.AppendBreak(b.B)

		return b, 0
	}()

	var e Encoder

	enc, err := e.Encode(nil, b, root)
	assertNoError(tb, err)

	exp := `"one_two_three"`
	if !bytes.Equal([]byte(exp), enc) {
		tb.Errorf(`wanted (%s), got %s`, exp, enc)
	}

	//	if tb.Failed() {
	tb.Logf("res %x\n%s", root, hex.Dump(b.B))
	// }
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
