package jq

import (
	"io"

	"github.com/nikandfor/errors"
	"github.com/nikandfor/jq/eval"
)

type (
	any = interface{}

	PipeParser struct {
		Pre  any
		Post any
	}

	Dot struct {
		io.Writer
	}
)

var JQ = eval.LeftToRight{
	New: PipeNewer,
	Op:  eval.Const("|"),
	Of:  Dot{},
}

func Parse(q string) (any, error) {
	return eval.ParseString(JQ, q)
}

func Build(w io.Writer, x any) (io.Writer, error) {
	y, err := eval.Build(w, x)
	if err != nil {
		return nil, err
	}

	_ = w

	w, ok := y.(io.Writer)
	if !ok {
		return nil, errors.New("not a writer")
	}

	return w, nil
}

func Compile(w io.Writer, q string) (io.Writer, error) {
	b, err := Parse(q)
	if err != nil {
		return nil, errors.Wrap(err, "parse")
	}

	w, err = Build(w, b)
	if err != nil {
		return nil, errors.Wrap(err, "build")
	}

	return w, nil
}

func PipeNewer(op, l, r any) any {
	return PipeParser{
		Pre:  l,
		Post: r,
	}
}

func (Dot) Parse(p []byte, st int) (any, int, error) {
	if st == len(p) || p[st] != '.' {
		return nil, st, errors.New("dot expected")
	}

	return Dot{}, st + 1, nil
}

func (Dot) Build(x any) (any, error) {
	w, ok := x.(io.Writer)
	if !ok {
		return nil, errors.New("arg is not io.Writer")
	}

	return Dot{w}, nil
}
