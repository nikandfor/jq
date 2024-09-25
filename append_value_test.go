package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestBufferAppendValue(tb *testing.T) {
	b := NewBuffer()

	off := b.AppendValue(Arr{
		Obj{"a", "b"},
		1, 2,
		[]byte{1, 2, 3},
		nil, Off(Null),
		Off(Zero),
		Raw{byte(cbor.String) | 3, 'a', 'b', 'c'},
		false,
		1.0,
		-1.0,
	})

	tb.Logf("buffer  %x\n%s", off, b.Dump())
}
