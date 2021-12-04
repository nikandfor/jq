package eval

import (
	"bytes"

	"github.com/nikandfor/errors"
)

type (
	Spaces uint64

	Ident []byte

	FirstOf []Parser

	AllOf []Parser

	Optional struct {
		Parser
	}

	Spacer struct {
		Spaces
		Of Parser
	}

	Const []byte

	None struct{}
)

var GoSpaces = NewSpaces(' ', '\n', '\t', '\r')

func SkipSpaces(p []byte, i int) int {
loop:
	for i < len(p) {
		switch p[i] {
		case ' ', '\n', '\t', '\r':
			i++
		default:
			break loop
		}
	}

	return i
}

func NewSpaces(skip ...byte) (ss Spaces) {
	for _, q := range skip {
		ss |= 1 << q
	}

	return
}

func (s Spaces) Skip(p []byte, i int) int {
	for i < len(p) {
		if s&(1<<p[i]) == 0 {
			break
		}

		i++
	}

	return i
}

func Spaced(p Parser, skip ...byte) Spacer {
	return Spacer{
		Spaces: NewSpaces(skip...),
		Of:     p,
	}
}

func (n Spacer) Parse(p []byte, st int) (x any, i int, err error) {
	st = n.Spaces.Skip(p, st)

	return n.Of.Parse(p, st)
}

func (n Const) Parse(p []byte, st int) (b any, i int, err error) {
	if bytes.HasPrefix(p[st:], n) {
		return n, st + len(n), nil
	}

	return nil, st, errors.New("expected %q", n)
}

func (n Ident) Parse(p []byte, st int) (b any, i int, err error) {
	if st == len(p) {
		return nil, st, errors.New("ident expected")
	}

	i = st

	c := p[i]

	switch {
	case c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_':
		i++
	default:
		return nil, st, errors.New("ident expected")
	}

loop:
	for i < len(p) {
		c := p[i]

		switch {
		case c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_':
			i++
		default:
			break loop
		}
	}

	return Ident(p[st:i]), i, nil
}

func (n FirstOf) Parse(p []byte, st int) (x any, i int, err error) {
	for _, r := range n {
		x, i, e := r.Parse(p, st)
		if e == nil {
			return x, i, nil
		}

		if err == nil {
			err = e
		}
	}

	return nil, st, err
}

func (n AllOf) Parse(p []byte, st int) (x any, i int, err error) {
	i = st

	res := make([]any, len(n))

	for i, r := range n {
		x, i, err = r.Parse(p, i)
		if err != nil {
			return nil, i, err
		}

		res[i] = x
	}

	return res, i, nil
}

func (n Optional) Parse(p []byte, st int) (x any, i int, err error) {
	x, i, err = n.Parser.Parse(p, st)
	if i == st {
		return None{}, st, nil
	}

	return
}
