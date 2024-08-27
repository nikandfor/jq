package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestEncoderDecoder(tb *testing.T) {
	var b []byte

	b, root := appendValBuf(b, 0, arr{true, false, code(True), code(False)})

	tb.Logf("buffer %x\n%s", root, Dump(b))

	var d Decoder

	exp := []int{0, 1, True, False}
	arr, _ := d.ArrayMap(b, 0, root, nil)

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

	arr = make([]int, len(exp))

	for i := range exp {
		_, arr[i] = d.ArrayMapIndex(b, 0, root, i)
	}

	cmp("elements")
}

func TestEncoderDecoderLong(tb *testing.T) {
	var e Encoder
	var b []byte

	b, el1 := appendValBuf(b, 0, true)
	b, el2 := appendValBuf(b, 0, false)

	b = append(b, cbor.String|cbor.Len2, 0x1, 0x00)
	b = append(b, make([]byte, 0x100)...)

	root := len(b)
	b = e.AppendArray(b, len(b), []int{el1, el2, True, False})

	tb.Logf("buffer %x\n%s", root, Dump(b))

	var d Decoder

	exp := []int{0, 1, True, False}
	arr, _ := d.ArrayMap(b, 0, root, nil)

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

	arr = make([]int, len(exp))

	for i := range exp {
		_, arr[i] = d.ArrayMapIndex(b, 0, root, i)
	}

	cmp("elements")
}
