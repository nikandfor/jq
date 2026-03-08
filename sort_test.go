package jq

import "testing"

func TestSortArrayNums(tb *testing.T) {
	b := NewBuffer()

	x := b.appendVal(arr{
		1.5,
		1,
		0,
		-1,
		-1.5,
	})

	e := b.appendVal(arr{
		-1.5,
		-1,
		0,
		1,
		1.5,
	})

	xarr := b.Reader().ArrayMap(x, nil)
	earr := b.Reader().ArrayMap(e, nil)

	defer func() {
		p := recover()
		if p == nil && !tb.Failed() {
			return
		}

		tb.Errorf("dump\n%s", b.Dump())

		if p != nil {
			panic(p)
		}
	}()

	b.SortArray(xarr)

	for i := range xarr {
		if !b.Equal(earr[i], xarr[i]) {
			tb.Errorf("%d: %+v != %+v", i, earr[i], xarr[i])
		}
	}

	tb.Logf("xarr %v", xarr)
}

func TestSortArray(tb *testing.T) {
	b := NewBuffer()

	x := b.appendVal(arr{
		obj{},
		arr{},
		string("b"),
		string("a"),
		[]byte("b"),
		[]byte("a"),
		1.5,
		1,
		0,
		-1,
		-1.5,
		true,
		false,
		nil,
	})

	e := b.appendVal(arr{
		nil,
		false,
		true,
		-1.5,
		-1,
		0,
		1,
		1.5,
		[]byte("a"),
		[]byte("b"),
		string("a"),
		string("b"),
		arr{},
		obj{},
	})

	xarr := b.Reader().ArrayMap(x, nil)
	earr := b.Reader().ArrayMap(e, nil)

	defer func() {
		p := recover()
		if p == nil && !tb.Failed() {
			return
		}

		tb.Errorf("dump\n%s", b.Dump())

		if p != nil {
			panic(p)
		}
	}()

	b.SortArray(xarr)

	for i := range xarr {
		if !b.Equal(earr[i], xarr[i]) {
			tb.Errorf("%d: %+v != %+v", i, earr[i], xarr[i])
		}
	}

	tb.Logf("xarr %v", xarr)
}

func TestSortMap(tb *testing.T) {
	b := NewBuffer()

	x := b.appendVal(obj{
		"b", 3,
		"d", 4,
		"b", 2,
		"a", 1,
		"b", -3,
	})

	e := b.appendVal(obj{
		"a", 1,
		"b", -3,
		"b", 2,
		"b", 3,
		"d", 4,
	})

	xobj := b.Reader().ArrayMap(x, nil)
	eobj := b.Reader().ArrayMap(e, nil)

	defer func() {
		p := recover()
		if p == nil && !tb.Failed() {
			return
		}

		tb.Errorf("dump\n%s", b.Dump())

		if p != nil {
			panic(p)
		}
	}()

	b.SortMap(xobj)

	for i := range xobj {
		if !b.Equal(eobj[i], xobj[i]) {
			tb.Errorf("%d: %+v != %+v", i, eobj[i], xobj[i])
		}
	}

	if tb.Failed() {
		tb.Logf("xobj %v", xobj)
	}
}
