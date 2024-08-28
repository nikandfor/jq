package jqjson

import (
	"bytes"
	"encoding/hex"
	"testing"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

func TestDecodeEncode(tb *testing.T) {
	data := []byte(`{"a":[{"q":"w","c":[1,2,3]},{"c":[4],"d":44}],"b":[{"c":[]},{"c":[5,6]}]}`)
	b := jq.NewBuffer(nil)

	var d Decoder

	off, i, err := d.Decode(b, data, 0)
	assertNoError(tb, err)
	assertEqual(tb, len(data), i)

	f := jq.NewArray(jq.NewIndex(jq.Iter{}, jq.Iter{}, "c", jq.Iter{}))

	res, _, err := f.ApplyTo(b, off, false)
	assertNoError(tb, err)

	var e Encoder

	enc, err := e.Encode(nil, b, res)
	assertNoError(tb, err)

	if !bytes.Equal([]byte(`[1,2,3,4,5,6]`), enc) {
		tb.Errorf("wanted array of 1 to 6, got %s", enc)
	}

	if tb.Failed() {
		tb.Logf("res %x -> %x\n%s", off, res, jq.DumpBuffer(b))
	}
}

func TestFilter(tb *testing.T) {
	data := `{"a":"b","c":"d"}`

	b := func() *jq.Buffer {
		var r []byte
		var e jq.Encoder

		r = e.CBOR.AppendTag(r, cbor.String, len(data))
		r = append(r, data...)

		return jq.NewBuffer(r)
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
		tb.Logf("hex R\n%s", hex.Dump(b.R))
		tb.Logf("hex W\n%s", hex.Dump(b.W))
		tb.Logf("buffer\n%s", jq.DumpBuffer(b))
	}
}

func TestMergeString(tb *testing.T) {
	b, root := func() (*jq.Buffer, int) {
		var r []byte
		var e jq.Encoder

		r = e.CBOR.AppendTag(r, cbor.String, -1)

		r = e.CBOR.AppendString(r, "one_")

		r = e.CBOR.AppendTag(r, cbor.String, -1)
		r = e.CBOR.AppendString(r, "two_")
		r = e.CBOR.AppendString(r, "three")
		r = e.CBOR.AppendBreak(r)

		r = e.CBOR.AppendBreak(r)

		return jq.NewBuffer(r), 0
	}()

	var e Encoder

	enc, err := e.Encode(nil, b, root)
	assertNoError(tb, err)

	exp := `"one_two_three"`
	if !bytes.Equal([]byte(exp), enc) {
		tb.Errorf(`wanted (%s), got %s`, exp, enc)
	}

	//	if tb.Failed() {
	tb.Logf("res %x\n%s", root, hex.Dump(b.R))
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
