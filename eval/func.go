package eval

import "github.com/nikandfor/errors"

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

	cl := n.Open + 1

	var l List

loop:
	for {
		var el any
		el, i, err = Optional{n.Of}.Parse(p, i)
		if err != nil {
			return nil, i, errors.Wrap(err, "list elem")
		}

		if el != (None{}) {
			l = append(l, el)
		}

		i = SkipSpaces(p, i)

		switch p[i] {
		case n.Sep:
			i++
		case cl:
			i++
			break loop
		}
	}

	return l, i, nil
}
