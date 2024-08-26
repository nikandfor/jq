package jq

import (
	"runtime"
	"testing"
)

func TestIndex(tb *testing.T) {
	d, root := appendValBuf(nil, obj{"a", 1, "b", obj{"c", arr{2, "3", obj{"d", 5}, true}}})

	//	log.Printf("data %x\n%s", root, Dump(d))

	b := NewBuffer(d)

	f := NewIndex("a")
	eoff := b.appendVal(1)

	off, more, err := f.ApplyTo(b, root, false)
	assertNoError(tb, err)
	assertEqualVal(tb, b, eoff, off)
	assertTrue(tb, !more)

	f = NewIndex("b", "c")
	eoff = b.appendVal(arr{2, "3", obj{"d", 5}, true})

	off, more, err = f.ApplyTo(b, root, false)
	assertNoError(tb, err)
	assertEqualVal(tb, b, eoff, off)
	assertTrue(tb, !more)

	f = NewIndex("b", "c", Iter{})
	testIter(tb, f, b, root, []any{2, "3", obj{"d", 5}, true})
}

func TestIndexIter(tb *testing.T) {
	d, root := appendValBuf(nil, arr{
		obj{"a", 1, "b", lab{lab: 4, val: 2}, "c", "d"},
		true,
	})

	//	log.Printf("data %x\n%s", root, Dump(d))

	b := NewBuffer(d)
	f := NewIndex(-2, Iter{})
	testIter(tb, f, b, root, []any{1, lab{lab: 4, val: 2}, "d"})
}

func TestIndexMultiIter(tb *testing.T) {
	d, root := appendValBuf(nil, arr{
		obj{"q", obj{"a", 1, "b", 2}},
		obj{"q", arr{}, "w", -5},
		obj{"q", arr{3, 4}},
	})

	//	log.Printf("data %x\n%s", root, Dump(d))

	b := NewBuffer(d)
	f := NewIndex(Iter{}, "q", Iter{})

	testIter(tb, f, b, root, []any{1, 2, 3, 4})
}

func testIter(tb testing.TB, f Filter, b *Buffer, root int, vals []any) {
	for j, elem := range vals {
		//	log.Printf("testIter  j %x  root %x", j, root)

		eoff := b.appendVal(elem)

		off, more, err := f.ApplyTo(b, root, j != 0)
		//	log.Printf("test iter  root %x  off %x  eoff %x  expect %v  err %v", root, off, eoff, elem, err)
		if assertNoError(tb, err, "j %d", j) {
			assertEqualVal(tb, b, eoff, off, "j %d  elem %v", j, elem)

			if j < len(vals)-1 {
				assertTrue(tb, more, "wanted more")
			}
		} else {
			return
		}
	}

	off, more, err := f.ApplyTo(b, root, true)
	if assertNoError(tb, err, "after") {
		assertEqualOff(tb, None, off, "after")
		assertTrue(tb, !more, "didn't want more")
	}

	if tb.Failed() {
		_, file, line, _ := runtime.Caller(1)
		tb.Logf("from %v:%d", file, line)
	}
}
