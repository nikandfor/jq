package jq

import "testing"

func TestIter(tb *testing.T) {
	b := NewBuffer(nil)
	root := b.appendVal(obj{"a", 1, "b", 2, "c", "d"})

	f := NewIter()

	for j, exp := range []any{1, 2, "d"} {
		eoff := b.appendVal(exp)

		off, more, err := f.ApplyTo(b, root, j != 0)
		if !assertNoError(tb, err) {
			break
		}
		if off == None {
			tb.Errorf("got None at step %d", j)
			break
		}

		assertEqualVal(tb, b, eoff, off)
		assertTrue(tb, more == (j < 2))
	}

	off, more, err := f.ApplyTo(b, root, true)
	assertNoError(tb, err)
	assertEqualOff(tb, None, off)
	assertTrue(tb, !more)
}
