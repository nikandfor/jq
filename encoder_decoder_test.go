package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestEncoderDecoder(tb *testing.T) {
	b := NewBuffer()

	root := b.appendVal(arr{
		raw{byte(cbor.Simple | cbor.True)},
		raw{byte(cbor.Simple | cbor.False)},
		True,
		False,
	})

	tb.Logf("buffer %x\n%s", root, b.Dump())

	var d Decoder

	exp := []Off{0, 1, True, False}
	arr, _ := d.ArrayMap(b.B, int(root), nil)

	cmp := func(n string) {
		if len(exp) != len(arr) {
			tb.Errorf("%s: wanted %x, got %x", n, exp, arr)
		} else {
			for i := range exp {
				if exp[i] != arr[i] {
					tb.Errorf("%s: i %x: wanted %x, got %x", n, i, exp, arr)
				}
			}
		}
	}

	cmp("array")

	arr = make([]Off, len(exp))

	for i := range exp {
		_, arr[i] = d.ArrayMapIndex(b.B, int(root), i)
	}

	cmp("elements")
}

func TestEncoderDecoderLong(tb *testing.T) {
	b := NewBuffer()

	el1 := b.appendVal(raw{byte(cbor.Simple | cbor.True)})
	el2 := b.appendVal(raw{byte(cbor.Simple | cbor.False)})

	b.B = append(b.B, byte(cbor.String|cbor.Len2), 0x1, 0x00)
	b.B = append(b.B, make([]byte, 0x100)...)

	root := len(b.B)
	b.B = b.Encoder.AppendArray(b.B, Off(len(b.B)), []Off{el1, el2, True, False})

	tb.Logf("buffer %x\n%s", root, b.Dump())

	var d Decoder

	exp := []Off{0, 1, True, False}
	arr, _ := d.ArrayMap(b.B, root, nil)

	cmp := func(n string) {
		if len(exp) != len(arr) {
			tb.Errorf("%s: wanted %x, got %x", n, exp, arr)
		} else {
			for i := range exp {
				if exp[i] != arr[i] {
					tb.Errorf("%s: i %x: wanted %x, got %x", n, i, exp, arr)
				}
			}
		}
	}

	cmp("array")

	arr = make([]Off, len(exp))

	for i := range exp {
		_, arr[i] = d.ArrayMapIndex(b.B, root, i)
	}

	cmp("elements")
}
