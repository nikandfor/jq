package eval

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/nikandfor/errors"
)

type (
	Spaces uint64

	Ident []byte

	AnyOf []Parser

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
	vst := n.Spaces.Skip(p, st)

	x, i, err = n.Of.Parse(p, vst)
	if err != nil {
		if i == vst {
			i = st
		}
	}

	return
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

func (n AnyOf) Parse(p []byte, st int) (x any, i int, err error) {
	for _, r := range n {
		x, i, err = r.Parse(p, st)
		if err == nil || i != st {
			return
		}
	}

	return nil, st, errors.New("expected one of %v", joinHuman(", ", " or ", []Parser(n)...))
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

func joinHuman(sep, last string, l ...Parser) string {
	var b strings.Builder

	for i, x := range l {
		if i+1 == len(l) {
			b.WriteString(last)
		} else if i != 0 {
			b.WriteString(sep)
		}

		fmt.Fprintf(&b, "%v", x)
	}

	return b.String()
}
