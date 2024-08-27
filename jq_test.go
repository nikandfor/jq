package jq

import (
	"bytes"
	"errors"
	"reflect"
	"runtime"
	"testing"

	"nikand.dev/go/cbor"
)

type (
	code int
	arr  []any
	obj  []any
	raw  []byte
	lab  struct {
		lab int
		val any
	}
)

func (b *Buffer) appendVal(v any) (off int) {
	b.W, off = appendValBuf(b.W, len(b.R), v)
	if off < 0 {
		return off
	}

	return off
}

func appendValBuf(w []byte, base int, v any) ([]byte, int) {
	var e Encoder
	var a []int
	var tag byte
	var lst []any

	off := base + len(w)

	switch v := v.(type) {
	case code:
		return w, int(v)
	case nil:
		return w, Null
	case raw:
		return append(w, v...), off
	case int:
		switch v {
		case 0:
			return w, Zero
		case 1:
			return w, One
		}

		w = e.CBOR.AppendInt(w, v)
		return w, off
	case string:
		w = e.CBOR.AppendString(w, v)
		return w, off
	case bool:
		if v {
			return w, True
		} else {
			return w, False
		}
	//	w = e.CBOR.AppendBool(w, v)
	//	return w, off
	case lab:
		w = e.CBOR.AppendLabeled(w, v.lab)
		w, _ = appendValBuf(w, base, v.val)
		return w, off
	case arr:
		tag = cbor.Array
		lst = []any(v)
	case obj:
		tag = cbor.Map
		lst = []any(v)
	default:
		panic(v)
	}

	for _, v := range lst {
		w, off = appendValBuf(w, base, v)
		a = append(a, off)
	}

	off = base + len(w)
	w = e.AppendArrayMap(w, tag, off, a)

	return w, off
}

func testError(tb testing.TB, f Filter, b *Buffer, root int, experr error) {
	tb.Logf("filter: %v", f)

	_, more, err := f.ApplyTo(b, root, false)
	assertErrorIs(tb, err, experr)
	assertTrue(tb, !more, "didn't want more")

	if tb.Failed() {
		_, file, line, _ := runtime.Caller(1)
		tb.Logf("from %v:%d", file, line)
	}
}

func testOne(tb testing.TB, f Filter, b *Buffer, root int, val any) {
	tb.Logf("filter: %v", f)

	eoff := b.appendVal(val)

	off, more, err := f.ApplyTo(b, root, false)
	assertNoError(tb, err)
	assertEqualVal(tb, b, eoff, off, "wanted %v", val)
	assertTrue(tb, !more, "didn't want more")

	if tb.Failed() {
		_, file, line, _ := runtime.Caller(1)
		tb.Logf("from %v:%d", file, line)
	}
}

func testIter(tb testing.TB, f Filter, b *Buffer, root int, vals []any) {
	tb.Logf("filter: %v", f)

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

func assertTrue(tb testing.TB, val bool, args ...any) bool {
	tb.Helper()

	if val {
		return true
	}

	tb.Errorf("Assertion failed: false")

	if len(args) != 0 {
		msg := args[0].(string)
		tb.Errorf(msg, args[1:]...)
	}

	return false
}

func assertErrorIs(tb testing.TB, err, target error, args ...any) bool {
	tb.Helper()

	if errors.Is(err, target) {
		return true
	}

	tb.Errorf("Assertion failed: error: %v wanted to be %v", err, target)

	if len(args) != 0 {
		msg := args[0].(string)
		tb.Errorf(msg, args[1:]...)
	}

	return false
}

func assertNoError(tb testing.TB, err error, args ...any) bool {
	tb.Helper()

	if err == nil {
		return true
	}

	tb.Errorf("Assertion failed: error: %v", err)

	if len(args) != 0 {
		msg := args[0].(string)
		tb.Errorf(msg, args[1:]...)
	}

	return false
}

func assertEqualOff(tb testing.TB, exp, val int, args ...any) bool {
	tb.Helper()

	if reflect.DeepEqual(exp, val) {
		return true
	}

	tb.Errorf("Assertion failed: %x is not equal to %x", exp, val)

	if len(args) != 0 {
		msg := args[0].(string)
		tb.Errorf(msg, args[1:]...)
	}

	return false
}

func assertEqualVal(tb testing.TB, b *Buffer, loff int, roff int, args ...any) bool {
	tb.Helper()

	if loff < 0 && roff < 0 && loff != roff {
		tb.Errorf("Assertion failed: %d (%#[1]x) != %d (%#[2]x)", loff, roff)

		return false
	}

	if b.Equal(loff, roff) {
		return true
	}

	var log bytes.Buffer

	(&Dumper{
		Writer: &log,
	}).ApplyTo(b, 0, false)

	tb.Errorf("Assertion failed: %x is not equal to %x, buffer:\n%s", loff, roff, log.Bytes())

	if len(args) != 0 {
		msg := args[0].(string)
		tb.Errorf(msg, args[1:]...)
	}

	return false
}
