package jq

import (
	"testing"
)

func TestConvert(tb *testing.T) {
	b := NewBuffer()

	r := b.appendVal(arr{
		3, 3.3, -4.4, "qwe", []byte("asd"),
	})

	testOne(tb, NewArray(NewQuery(Iter{}, NewConvert(ToString))), b, r, arr{"3", "3.3", "-4.4", "qwe", "asd"})
	testOne(tb, NewArray(NewQuery(Iter{}, NewConvert(ToBytes))), b, r, arr{[]byte("3"), []byte("3.3"), []byte("-4.4"), []byte("qwe"), []byte("asd")})

	r = b.appendVal(arr{
		3, 3.3, -4.4, "45", []byte("123"),
	})

	testOne(tb, NewArray(NewQuery(Iter{}, NewConvert(ToInt))), b, r, arr{3, 3, -4, 45, 123})

	r = b.appendVal(arr{
		3, -5, 3.3, -4.4, "45", []byte("12.3"),
	})

	testOne(tb, NewArray(NewQuery(Iter{}, NewConvert(ToFloat))), b, r, arr{3., -5., 3.3, -4.4, 45., 12.3})
}
