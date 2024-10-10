package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestTypeMatch(tb *testing.T) {
	b := NewBuffer()
	r := b.appendVal(arr{1, -1, 1.1, false, "a", arr{}, obj{}})

	testOne(tb, NewMap(NewTypeMatch(cbor.Int)), b, r,
		arr{True, False, False, False, False, False, False})

	testOne(tb, NewMap(NewTypeMatch(cbor.Neg)), b, r,
		arr{False, True, False, False, False, False, False})

	testOne(tb, NewMap(NewTypeMatch(cbor.Int, cbor.Neg)), b, r,
		arr{True, True, False, False, False, False, False})

	testOne(tb, NewMap(NewTypeMatch(cbor.Simple|cbor.Float64)), b, r,
		arr{False, False, True, False, False, False, False})

	testOne(tb, NewMap(NewTypeMatch(cbor.Simple|cbor.True, cbor.Simple|cbor.False)), b, r,
		arr{False, False, False, True, False, False, False})

	testOne(tb, NewMap(NewTypeMatch(cbor.Bytes, cbor.String)), b, r,
		arr{False, False, False, False, True, False, False})

	testOne(tb, NewMap(NewTypeMatch(cbor.Array)), b, r,
		arr{False, False, False, False, False, True, False})

	testOne(tb, NewMap(NewTypeMatch(cbor.Map)), b, r,
		arr{False, False, False, False, False, False, True})
}
