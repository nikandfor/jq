package jqbase64

import (
	"bytes"
	"testing"

	"nikand.dev/go/jq"
)

func TestEncodeDecode(tb *testing.T) {
	b := jq.NewBuffer()

	for _, data := range []string{``, `cm9vdCAxICAgZmlsdGVyOiAoLltdIHwgaWYgLiB0aGVuIHsieCI6IC59IGVsc2UgIm5vIiBlbmQp`} {
		b.Reset()

		var d Decoder

		off, i, err := d.Decode(b, []byte(data), 0)
		assertNoError(tb, err)
		assertEqual(tb, len(data), i)

		var e Encoder

		enc, err := e.Encode(nil, b, off)
		assertNoError(tb, err)

		if !bytes.Equal([]byte(data), enc) {
			tb.Errorf("wanted array of 1 to 6, got %s", enc)
		}

		if tb.Failed() {
			tb.Logf("res %x -> %x\n%s", off, off, jq.Dump(b))
			break
		}
	}
}

func TestFilter(tb *testing.T) {
	b := jq.NewBuffer()
	d := b.AppendValue(`cm9vdCAxICAgZmlsdGVyOiAoLltdIHwgaWYgLiB0aGVuIHsieCI6IC59IGVsc2UgIm5vIiBlbmQp`)

	f := jq.NewPipe(
		NewDecoder(nil),
		NewEncoder(nil),
	)

	res, _, err := f.ApplyTo(b, d, false)
	assertNoError(tb, err)

	if !b.Equal(d, res) {
		tb.Errorf("not equal (%s) != (%s)", d, res)
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
