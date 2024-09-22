package jq

import (
	"bytes"
	"errors"
	"reflect"
	"runtime"
	"testing"
)

func testError(tb testing.TB, f Filter, b *Buffer, root Off, experr error) {
	tb.Logf("filter: %v", f)

	_, more, err := f.ApplyTo(b, root, false)
	assertErrorIs(tb, err, experr)
	assertTrue(tb, !more, "didn't want more")

	if tb.Failed() {
		_, file, line, _ := runtime.Caller(1)
		tb.Logf("from %v:%d", file, line)
	}
}

func testSame(tb testing.TB, f Filter, b *Buffer, eoff, off Off) {
	tb.Logf("root %v   filter: %v", off, f)

	off, more, err := f.ApplyTo(b, off, false)
	assertNoError(tb, err)
	assertEqualOff(tb, eoff, off)
	assertTrue(tb, !more, "didn't want more")

	if tb.Failed() {
		_, file, line, _ := runtime.Caller(1)
		tb.Logf("from %v:%d", file, line)
	}
}

func testOne(tb testing.TB, f Filter, b *Buffer, root Off, val any) {
	tb.Logf("root %v   filter: %v", root, f)

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

func testOnePath(tb testing.TB, f FilterPath, b *Buffer, root Off, val any, exp NodePath) {
	tb.Logf("root %v   filter: %v", root, f)

	eoff := b.appendVal(val)

	off, path, more, err := f.ApplyToGetPath(b, root, nil, false)
	assertNoError(tb, err)
	assertEqualVal(tb, b, eoff, off, "wanted %v", val)
	assertDeepEqual(tb, exp, path)
	assertTrue(tb, !more, "didn't want more")

	if tb.Failed() {
		_, file, line, _ := runtime.Caller(1)
		tb.Logf("from %v:%d", file, line)
	}
}

func testIter(tb testing.TB, f Filter, b *Buffer, root Off, vals []any) {
	tb.Logf("root %v   filter: %v", root, f)

	defer func() {
		p := recover()
		if p == nil {
			return
		}

		defer panic(p)

		tb.Logf("buffer  root %v\n%s", root, Dump(b))
	}()

	for j, elem := range vals {
		eoff := b.appendVal(elem)

		off, more, err := f.ApplyTo(b, root, j != 0)
		if assertNoError(tb, err, "j %d", j) {
			assertEqualVal(tb, b, eoff, off, "j %d  value %v", j, elem)

			assertTrue(tb, more == (j+1 < len(vals)), "wanted more: %v", j+1 < len(vals))
		} else {
			return
		}
	}

	off, more, err := f.ApplyTo(b, root, len(vals) != 0)
	if assertNoError(tb, err, "after") {
		assertEqualOff(tb, None, off, "after")
		assertTrue(tb, !more, "didn't want more")
	}

	if tb.Failed() {
		_, file, line, _ := runtime.Caller(1)
		tb.Logf("from %v:%d", file, line)
	}
}

func testIterPath(tb testing.TB, f FilterPath, b *Buffer, root Off, vals []any, paths []NodePath) {
	tb.Logf("root %v   filter: %v", root, f)

	defer func() {
		p := recover()
		if p == nil {
			return
		}

		defer panic(p)

		tb.Logf("buffer  root %v\n%s", root, Dump(b))
	}()

	var base NodePath

	for j, elem := range vals {
		eoff := b.appendVal(elem)

		off, path, more, err := f.ApplyToGetPath(b, root, base, j != 0)
		if assertNoError(tb, err, "j %d", j) {
			assertEqualVal(tb, b, eoff, off, "j %d  value %v", j, elem)
			assertDeepEqual(tb, paths[j], path, "wanted path %v", paths[j])

			assertTrue(tb, more == (j+1 < len(vals)), "wanted more: %v", j+1 < len(vals))
		} else {
			return
		}
	}

	off, path, more, err := f.ApplyToGetPath(b, root, base, len(vals) != 0)
	if assertNoError(tb, err, "after") {
		assertEqualOff(tb, None, off, "after")
		assertDeepEqual(tb, base, path, "wanted path %v", base)

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

func assertEqualOff(tb testing.TB, exp, val Off, args ...any) bool {
	tb.Helper()

	if exp == val {
		return true
	}

	tb.Errorf("Assertion failed: %x is not equal to %x", exp, val)

	if len(args) != 0 {
		msg := args[0].(string)
		tb.Errorf(msg, args[1:]...)
	}

	return false
}

func assertEqualVal(tb testing.TB, b *Buffer, loff, roff Off, args ...any) bool {
	tb.Helper()

	if b.Equal(loff, roff) {
		return true
	}

	if loff < 0 && roff < 0 && loff != roff {
		tb.Errorf("Assertion failed: %v != %v", loff, roff)
	} else {

		var log bytes.Buffer

		(&Dumper{
			Writer: &log,
		}).ApplyTo(b, 0, false)

		tb.Errorf("Assertion failed: %v is not equal to %v, buffer:\n%s", loff, roff, log.Bytes())
	}

	if len(args) != 0 {
		msg := args[0].(string)
		tb.Errorf(msg, args[1:]...)
	}

	return false
}

func assertDeepEqual(tb testing.TB, exp, got any, args ...any) bool {
	tb.Helper()

	if reflect.DeepEqual(exp, got) {
		return true
	}

	tb.Errorf("expected to be equal\nexp: %#v\ngot: %#v", exp, got)

	if len(args) != 0 {
		msg := args[0].(string)
		tb.Errorf(msg, args[1:]...)
	}

	return false
}
