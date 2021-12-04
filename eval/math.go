package eval

import (
	"reflect"
	"strconv"

	"github.com/nikandfor/errors"
)

type (
	Math struct{}

	LeftToRight struct {
		Op  Parser
		Of  Parser
		New func(op, l, r any) any
	}

	BinOp struct {
		Op   any
		L, R any
	}

	Num struct{}

	Parentheses struct {
		Open byte

		Of Parser
	}

	Int []byte

	Float []byte
)

func (n Math) Parse(p []byte, st int) (any, int, error) {
	return LeftToRight{
		New: BinOpNewer,
		Op:  FirstOf{Const{'+'}, Const{'-'}},
		Of: LeftToRight{
			New: BinOpNewer,
			Op:  FirstOf{Const{'*'}, Const{'/'}},
			Of: Spaced(
				FirstOf{
					Ident(""),
					Num{},
					Parentheses{
						Open: '(',
						Of:   Math{},
					},
				},
				' '),
		},
	}.Parse(p, st)
}

func (n LeftToRight) Parse(p []byte, st int) (l any, i int, err error) {
	l, i, err = n.Of.Parse(p, st)
	if err != nil {
		return
	}

	for i < len(p) {
		i = SkipSpaces(p, i)

		var op any
		op, i, err = Optional{n.Op}.Parse(p, i)
		if err != nil {
			return
		}

		if op == (None{}) {
			return l, i, nil
		}

		var r any
		r, i, err = n.Of.Parse(p, i)
		if err != nil {
			return
		}

		l = n.New(op, l, r)
	}

	return
}

func BinOpNewer(op, l, r any) any {
	return BinOp{
		Op: op,
		L:  l,
		R:  r,
	}
}

func (n Parentheses) Parse(p []byte, st int) (x any, i int, err error) {
	i = SkipSpaces(p, st)

	if p[i] != n.Open {
		return nil, st, errors.New("%q expected", n.Open)
	}

	i++

	x, i, err = n.Of.Parse(p, i)
	if err != nil {
		return
	}

	i = SkipSpaces(p, i)

	if p[i] != n.Open+1 {
		return nil, i, errors.New("%q expected", n.Open+1)
	}

	i++

	return x, i, nil
}

func (n Num) Parse(p []byte, st int) (x any, i int, err error) {
	i = st

	dot := false

loop:
	for ; i < len(p); i++ {
		c := p[i]

		switch {
		case c >= '0' && c <= '9':
		case !dot && c == '.':
			dot = true
		default:
			break loop
		}
	}

	if i == st || i == st+1 && p[st] == '.' {
		return nil, st, errors.New("num expected")
	}

	if dot {
		x = Float(p[st:i])
	} else {
		x = Int(p[st:i])
	}

	return
}

func (op BinOp) Build() (x any, err error) {
	s, ok := op.Op.(Const)
	if !ok {
		return nil, errors.New("unsupported op: %q", op.Op)
	}

	l, err := Build(op.L)
	if err != nil {
		return nil, errors.Wrap(err, "build left")
	}

	r, err := Build(op.R)
	if err != nil {
		return nil, errors.Wrap(err, "build right")
	}

	if IsConst(op.L) && IsConst(op.R) {
		x, err = op.Eval(nil, string(s), l, r)
		if err != nil {
			return nil, errors.Wrap(err, "eval const")
		}

		return x, nil
	}

	return func(ctx any) (any, error) {
		return op.Eval(ctx, string(s), l, r)
	}, nil
}

func (BinOp) Eval(ctx any, op string, l, r any) (x any, err error) {
	l, r, err = ToTheSameType(ctx, l, r)
	if err != nil {
		return nil, errors.Wrap(err, "invalid operation: %v", op)
	}

	switch op {
	case "+":
		switch l := l.(type) {
		case int64:
			r := r.(int64)

			return l + r, nil
		case float64:
			r := r.(float64)

			return l + r, nil
		default:
			return nil, errors.New("unsupported type: %T", l)
		}
	case "*":
		switch l := l.(type) {
		case int64:
			r := r.(int64)

			return l * r, nil
		case float64:
			r := r.(float64)

			return l * r, nil
		default:
			return nil, errors.New("unsupported type: %T", l)
		}
	default:
		return nil, errors.New("unsupported op: %q", op)
	}
}

func ToTheSameType(ctx, l, r any) (lc, rc any, err error) {
	l, err = Eval(ctx, l)
	if err != nil {
		return nil, nil, errors.Wrap(err, "eval left")
	}

	r, err = Eval(ctx, r)
	if err != nil {
		return nil, nil, errors.Wrap(err, "eval left")
	}

	lt := reflect.TypeOf(l)
	rv := reflect.ValueOf(r)

	if lt == nil {
		return nil, nil, errors.New("left: invalid value")
	}

	if !rv.IsValid() {
		return nil, nil, errors.New("right: invalid value")
	}

	if lt == rv.Type() {
		return l, r, nil
	}

	return nil, nil, errors.New("mismatched types %v and %v", lt, rv.Type())

	/*
		if !rv.CanConvert(lt) {
			return nil, nil, errors.New("mismatched types %v and %v", lt, rv.Type())
		}

		r = rv.Convert(lt).Interface()

		return l, r, nil
	*/
}

func (x Int) Build() (y any, err error) {
	return strconv.ParseInt(string(x), 10, 64)
}

func (x Float) Build() (y any, err error) {
	return strconv.ParseFloat(string(x), 64)
}
