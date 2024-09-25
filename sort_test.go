package jq

import "testing"

func TestSortCompare(tb *testing.T) {
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
}
