package jq

import (
	"bytes"
	"reflect"
	"testing"

	"nikand.dev/go/cbor"
)

type (
	code int
	arr  []any
	obj  []any
	lab  struct {
		lab int
		val any
	}
)

func (b *Buffer) appendVal(v any) (off int) {
	b.W, off = appendValBuf(b.W, v)
	if off < 0 {
		return off
	}

	return len(b.R) + off
}

func appendValBuf(w []byte, v any) ([]byte, int) {
	var e Encoder
	var a []int
	var tag byte
	var lst []any

	off := len(w)

	switch v := v.(type) {
	case code:
		return w, int(v)
	case int:
		w = e.CBOR.AppendInt(w, v)
		return w, off
	case string:
		w = e.CBOR.AppendString(w, v)
		return w, off
	case nil:
		w = e.CBOR.AppendNull(w)
		return w, off
	case bool:
		w = e.CBOR.AppendBool(w, v)
		return w, off
	case lab:
		w = e.CBOR.AppendLabeled(w, v.lab)
		w, _ = appendValBuf(w, v.val)
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
		w, off = appendValBuf(w, v)
		a = append(a, off)
	}

	off = len(w)
	w = e.AppendArrayMap(w, tag, off, a)

	return w, off
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
