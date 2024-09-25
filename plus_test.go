package jq

import (
	"testing"

	"nikand.dev/go/cbor"
)

func TestPlusInt(tb *testing.T) {
	b := NewBuffer()
	six := b.appendVal(6)
	minusfive := b.appendVal(-5)

	testOne(tb, NewPlus(Zero, Zero), b, Null, Zero)
	testOne(tb, NewPlus(Zero, One), b, Null, One)
	testOne(tb, NewPlus(One, Zero), b, Null, One)
	testOne(tb, NewPlus(One, Null), b, Null, One)
	testOne(tb, NewPlus(Null, One), b, Null, One)
	testOne(tb, NewPlus(Null, Null), b, Null, Null)
	testOne(tb, NewPlus(six, six), b, Null, 12)
	testOne(tb, NewPlus(six, minusfive), b, Null, One)
	testOne(tb, NewPlus(minusfive, minusfive), b, Null, -10)
}

func TestPlusString(tb *testing.T) {
	b := NewBuffer()
	a := b.appendVal("aa")
	c := b.appendVal("ccc")

	testOne(tb, NewPlus(a, c), b, Null, "aaccc")
	testOne(tb, NewPlus(c, a), b, Null, "cccaa")
	testOne(tb, NewPlus(a, EmptyString), b, Null, "aa")
	testOne(tb, NewPlus(EmptyString, c), b, Null, "ccc")
	testOne(tb, NewPlus(EmptyString, EmptyString), b, Null, "")
	testOne(tb, NewPlus(a, Null), b, Null, "aa")
	testOne(tb, NewPlus(Null, c), b, Null, "ccc")
}

func TestPlusArray(tb *testing.T) {
	b := NewBuffer()
	a := b.appendVal(arr{1, 2, 3})
	c := b.appendVal(arr{4, 5})

	testOne(tb, NewPlus(a, c), b, Null, arr{1, 2, 3, 4, 5})
	testOne(tb, NewPlus(c, a), b, Null, arr{4, 5, 1, 2, 3})
	testOne(tb, NewPlus(a, EmptyArray), b, Null, arr{1, 2, 3})
	testOne(tb, NewPlus(EmptyArray, c), b, Null, arr{4, 5})
	testOne(tb, NewPlus(EmptyArray, EmptyArray), b, Null, arr{})
	testOne(tb, NewPlus(a, Null), b, Null, a)
	testOne(tb, NewPlus(Null, c), b, Null, c)
}

func TestPlusMap(tb *testing.T) {
	b := NewBuffer()
	e := b.appendVal(obj{})
	a := b.appendVal(obj{"a", 1, "b", 2})
	c := b.appendVal(obj{"a", 3, "c", 4})
	d := b.appendVal(obj{"d", 10})

	testOne(tb, NewPlus(a, a), b, Null, a)
	testOne(tb, NewPlus(a, e), b, Null, a)
	testOne(tb, NewPlus(e, a), b, Null, a)
	testOne(tb, NewPlus(a, c), b, Null, obj{"a", 3, "b", 2, "c", 4})
	testOne(tb, NewPlus(c, a), b, Null, obj{"a", 1, "c", 4, "b", 2})
	testOne(tb, NewPlus(a, d), b, Null, obj{"a", 1, "b", 2, "d", 10})
}

func TestPlusFloat(tb *testing.T) {
	b := NewBuffer()
	a := b.appendVal(6.5)
	c := b.appendVal(-5.25)
	d := b.appendVal(4)
	e := b.appendVal(-3)

	testOne(tb, NewPlus(a, a), b, Null, 6.5+6.5)
	testOne(tb, NewPlus(a, c), b, Null, 6.5-5.25)
	testOne(tb, NewPlus(a, d), b, Null, 6.5+4)
	testOne(tb, NewPlus(a, e), b, Null, 6.5-3)
	testOne(tb, NewPlus(e, c), b, Null, -3-5.25)
}

func TestPlusIter(tb *testing.T) {
	b := NewBuffer()
	f := b.appendVal(5)
	x := b.appendVal(arr{1, 2})
	y := b.appendVal(arr{10, 20})

	testIter(tb, NewPlus(NewIterOf(x), f), b, Null, []any{6, 7})
	testIter(tb, NewPlus(f, NewIterOf(y)), b, Null, []any{15, 25})
	testIter(tb, NewPlus(NewIterOf(x), NewIterOf(y)), b, Null, []any{11, 21, 12, 22})
}

func TestPlusError(tb *testing.T) {
	b := NewBuffer()

	testError(tb, NewPlus(One, EmptyString), b, Null, PlusError(cbor.Int)|PlusError(cbor.String)<<8)
}
