package jqcsv

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
)

func TestDecodeEncode(tb *testing.T) {
	b := jq.NewBuffer()

	var d Decoder
	var e Encoder

	for _, x := range []string{
		"",
		`1`, `10`, `-1`, `-100`,
		//	`"1"`, `"10"`, `"-1"`, `"-100"`,
		`1,2,3`,
		`a,b,c`,
		` a b,c d,e `,
		`"qwe""qwe""q"`,
	} {
		b.Reset()
		d.Reset()
		e.Reset()

		log.Printf("running test: %q", x)

		off, i, err := d.Decode(b, []byte(x), 0)
		assertNoError(tb, err)
		assertEqual(tb, len(x), i)

		if x == "" && off == jq.None {
			continue
		}

		if off == jq.None {
			tb.Errorf("decoded to None: (%s)", x)
			break
		}

		y, err := e.Encode(nil, b, off)
		assertNoError(tb, err)

		assertBytes(tb, []byte(x), bytes.TrimSuffix(y, []byte{'\n'}))

		if tb.Failed() {
			break
		}
	}

	if tb.Failed() {
		tb.Logf("dump\n%s", b.Dump())
	}
}

func TestMap(tb *testing.T) {
	x := `a,b,c,d
1,2,3,4
q,w,e,r
,,true,false
`

	var buf bytes.Buffer
	dump := jq.NewDumper(&buf)

	d := NewDecoder()
	d.Header = true
	d.Tag = cbor.Map

	e := NewEncoder()
	e.MapHeader = true

	s := jq.NewSandwich(d, e)

	y, err := s.ProcessAll(dump, nil, []byte(x))
	assertNoError(tb, err)

	assertBytes(tb, []byte(strings.TrimPrefix(x, "\n")), y)

	// tb.Logf("dump\n%s", y)
	// tb.Logf("dump\n%s", buf.String())
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
