package eval

import "github.com/nikandfor/errors"

type (
	Builder interface {
		Build() (any, error)
	}

	Conster interface {
		Const() bool
	}
)

func IsConst(x any) bool {
	if c, ok := x.(Conster); ok {
		return c.Const()
	}

	switch x.(type) {
	case Const, Int, Float:
		return true
	}

	return false
}

func Build(x any) (y any, err error) {
	if b, ok := x.(Builder); ok {
		y, err = b.Build()
		if err != nil {
			return nil, errors.Wrap(err, "%T", x)
		}

		return y, nil
	}

	switch x := x.(type) {
	case Ident:
		return func(ctx any) (y any, err error) {
			if v, ok := ctx.(Varer); ok {
				return v.Var(string(x))
			}

			if v, ok := ctx.(map[string]interface{}); ok {
				return v[string(x)], nil
			}

			return ctx, nil
		}, nil
	default:
		return nil, errors.New("unsupported type: %T", x)
	}
}
