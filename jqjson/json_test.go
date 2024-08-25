package jqjson

import (
	"bytes"
	"testing"

	"nikand.dev/go/jq"
)

func TestJSON(tb *testing.T) {
	data := []byte(`{"a":[{"q":"w","c":[1,2,3]},{"c":[4],"d":44}],"b":[{"c":[]},{"c":[5,6]}]}`)

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

	if !bytes.Equal([]byte(`[1,2,3,4,5,6]`), enc) {
		tb.Errorf("wanted array of 1 to 6, got %s", enc)
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
