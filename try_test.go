package jq

import (
	"errors"
	"testing"
)

func TestTryCatch(tb *testing.T) {
	b := NewBuffer()
	str := b.appendVal("str")

	testIter(tb, NewTry(NewQuery(str, Key("key")), nil), b, Null, []any{})
	testOne(tb, NewTry(ErrorText("some error"), Dot{}), b, Null, "some error")

	e := errors.New("some err")
	testError(tb, NewErrorExpr(NewErrorErr(e)), b, Null, e)

	testIter(tb, NewErrorExpr(NewIter()), b, EmptyArray, []any{})

	a := b.appendVal(arr{"first", "second"})
	f := NewErrorExpr(NewIter())
	tb.Logf("running %v", f)

	res, more, err := f.ApplyTo(b, a, false)
	if err == nil || err.Error() != "first" {
		tb.Errorf("expected error: first, got %v", err)
	}
	if more || res != None {
		tb.Errorf("unexpected result %v %v", res, more)
	}
}
