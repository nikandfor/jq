package eval

import (
	"fmt"

	"github.com/nikandfor/errors"
)

type (
	FuncCallParser struct {
		Ident Parser
		Args  Parser
	}

	FuncCall struct {
		Ident any
		Args  any
	}

	ListParser struct {
		Open byte
		Sep  byte

		Of Parser
	}

	List []any
)

func (n FuncCallParser) Parse(p []byte, st int) (x any, i int, err error) {
	var f FuncCall

	f.Ident, i, err = n.Ident.Parse(p, st)
	if err != nil {
		return x, i, errors.Wrap(err, "ident")
	}

	if n.Args != nil {
		f.Args, i, err = n.Args.Parse(p, i)
		if err != nil {
			return x, i, errors.Wrap(err, "args")
		}
	}

	return f, i, nil
}

func (n ListParser) Parse(p []byte, st int) (x any, i int, err error) {
	i = st

	if i == len(p) {
		return nil, st, errors.New("list expected")
	}

	if p[i] != n.Open {
		return nil, st, errors.New("%q expected", n.Open)
	}

	i++

	cl := Closing(n.Open)

	var l List

loop:
	for {
		i = SkipSpaces(p, i)

		if p[i] == cl {
			i++
			break loop
		} else if len(l) != 0 {
			if p[i] == n.Sep {
				i++
			} else {
				return nil, i, errors.New("%q, %q or %v expected", cl, n.Sep, n.Of)
			}
		}

		var el any
		el, i, err = n.Of.Parse(p, i)
		if err != nil {
			return nil, i, errors.Wrap(err, "list elem")
		}

		l = append(l, el)
	}

	return l, i, nil
}

func (x ListParser) String() string {
	return fmt.Sprintf("%c %v%c ... %c", x.Open, x.Of, x.Sep, Closing(x.Open))
}

func Closing(x byte) byte {
	switch x {
	case '(':
		return x + 1
	case '[', '{', '<':
		return x + 2
	default:
		return x
	}
}
