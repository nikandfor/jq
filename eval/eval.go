package eval

import "github.com/nikandfor/errors"

type (
	Evaluable interface {
		Eval(ctx any) (any, error)
	}

	Varer interface {
		Var(string) (any, error)
	}
)

func Eval(ctx, x any) (y any, err error) {
	switch x := x.(type) {
	case Evaluable:
		return x.Eval(ctx)
	case func(ctx any) (any, error):
		return x(ctx)
	case int64:
		return x, nil
	case float64:
		return x, nil
	}

	return nil, errors.New("unsupported type: %T", x)
}
