package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestBufferAppendValue(tb *testing.T) {
	b := NewBuffer(nil)

	off := b.AppendValue(Arr{
		Obj{"a", "b"},
		1, 2,
		[]byte{1, 2, 3},
		nil, Code(Null),
		Code(Zero),
		Raw{cbor.String | 3, 'a', 'b', 'c'},
		false,
	})

	tb.Logf("buffer  %x\n%s", off, DumpBuffer(b))
}
