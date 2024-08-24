package jq

import (
	"bytes"
	"reflect"
	"testing"

	"nikand.dev/go/cbor"
)

type (
	arr []any
	obj []any
	lab struct {
		lab int
		val any
	}
)

func (b *Buffer) appendVal(v any) int {
	var off int
	b.w, off = appendValBuf(b.w, v)
	return len(b.r) + off
}

func appendValBuf(w []byte, v any) ([]byte, int) {
	var e Encoder
	var a []int
	var lst []any

	off := len(w)

	switch v := v.(type) {
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
		lst = []any(v)
	case obj:
		lst = []any(v)
	default:
		panic(v)
	}

	for _, v := range lst {
		w, off = appendValBuf(w, v)
		a = append(a, off)
	}

	tag := byte(cbor.Array)
	if _, ok := v.(obj); ok {
		tag = cbor.Map
	}

	off = len(w)
	w = e.AppendArrayMap(w, tag, off, a)

	return w, off
}

func equal(b *Buffer, loff int, roff int) (res bool) {
	br := b.Reader()

	//	log.Printf("equal %x %x", loff, roff)
	//	defer func() { log.Printf("equal %x %x  =>  %v", loff, roff, res) }()

	if loff == roff {
		return true
	}

	tag := br.Tag(loff)
	rtag := br.Tag(roff)

	if tag != rtag {
		return false
	}

	switch tag {
	case cbor.Int, cbor.Neg, cbor.Bytes, cbor.String, cbor.Simple, cbor.Labeled:
		lraw := br.Raw(loff)
		rraw := br.Raw(roff)

		return bytes.Equal(lraw, rraw)
	case cbor.Array, cbor.Map:
	default:
		panic(tag)
	}

	larr := br.ArrayMap(loff, nil)
	rarr := br.ArrayMap(roff, nil)

	if len(larr) != len(rarr) {
		return false
	}

	for i := range larr {
		if !equal(b, larr[i], rarr[i]) {
			return false
		}
	}

	return true
}

func assertTrue(tb testing.TB, val bool, args ...any) bool {
	tb.Helper()

	if val {
		return true
	}

	tb.Errorf("false")

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

	tb.Errorf("error: %v", err)

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

	tb.Errorf("not equal: %x != %x", exp, val)

	if len(args) != 0 {
		msg := args[0].(string)
		tb.Errorf(msg, args[1:]...)
	}

	return false
}

func assertEqualVal(tb testing.TB, b *Buffer, loff int, roff int, args ...any) bool {
	tb.Helper()

	if loff != roff && loff < 0 || roff < 0 {
		tb.Errorf("none/nil != non-none/nil: %d (%#[1]x) %d (%#[2]x)", loff, roff)

		return false
	}

	if equal(b, loff, roff) {
		return true
	}

	var log bytes.Buffer

	(&Dumper{
		Writer: &log,
	}).ApplyTo(b, 0, false)

	tb.Errorf("not equal %x  %x at\n%s", loff, roff, log.Bytes())

	if len(args) != 0 {
		msg := args[0].(string)
		tb.Errorf(msg, args[1:]...)
	}

	return false
}
