package jq

import (
	"errors"
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Try struct {
		Expr  Filter
		Catch Filter

		err bool
	}

	ErrorText string

	ErrorExpr struct {
		Expr Filter
	}

	ErrorErr struct {
		Err error
	}
)

func NewTry(expr, catch Filter) *Try { return &Try{Expr: expr, Catch: catch} }

func (f *Try) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if !next {
		f.err = false
	}
	if f.err {
		return None, false, nil
	}

	e := csel[Filter](f.Expr != nil, f.Expr, Dot{})

	res, more, err = e.ApplyTo(b, off, next)
	if err == nil {
		return res, more, nil
	}

	f.err = true

	if f.Catch == nil {
		return None, false, nil
	}

	res = b.AppendValue(err.Error())

	return f.Catch.ApplyTo(b, res, false)
}

func NewErrorExpr(e Filter) ErrorExpr { return ErrorExpr{Expr: e} }

func (f ErrorExpr) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	if f.Expr == nil {
		return None, false, fmt.Errorf("error of nil expression")
	}

	res, _, err = f.Expr.ApplyTo(b, off, false)
	if err != nil {
		return res, false, err
	}
	if res == None {
		return None, false, nil
	}

	br := b.Reader()
	tag := br.Tag(res)

	switch tag {
	case cbor.Bytes, cbor.String:
		s := br.Bytes(res)

		return None, false, errors.New(string(s))
	}

	return None, false, fmt.Errorf("not a string error: %+v", res)
}

func NewErrorErr(err error) ErrorErr { return ErrorErr{Err: err} }

func (f ErrorErr) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	return None, false, f.Err
}

func (f ErrorText) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	return None, false, errors.New(string(f))
}

func (f Try) String() string {
	if f.Catch == nil {
		return fmt.Sprintf("try %+v", f.Expr)
	}

	return fmt.Sprintf("try %+v catch %+v", f.Expr, f.Catch)
}

func (f ErrorText) String() string { return fmt.Sprintf(`error(%q)`, string(f)) }
func (f ErrorExpr) String() string { return fmt.Sprintf(`error(%+v)`, f.Expr) }
func (f ErrorErr) String() string  { return fmt.Sprintf(`error(%q)`, f.Err) }
