package eval

import "github.com/nikandfor/errors"

type (
	any = interface{}

	Parser interface {
		Parse(p []byte, st int) (x any, i int, err error)
	}

	Stack struct {
		s []any
	}
)

func ParseString(root Parser, q string) (any, error) {
	return Parse(root, []byte(q))
}

func Parse(root Parser, p []byte) (x any, err error) {
	x, i, err := root.Parse(p, 0)
	if err != nil {
		return nil, errors.Wrap(err, "off: %x", i)
	}

	if i != len(p) {
		return nil, errors.New("not complete read: %x/%x", i, len(p))
	}

	return x, nil
}

func (st *Stack) Peek() any {
	if len(st.s) == 0 {
		return nil
	}

	return st.s[len(st.s)-1]
}

func (st *Stack) Pop() (x any) {
	if len(st.s) == 0 {
		panic("empty stack")
	}

	l := len(st.s) - 1

	x = st.s[l]

	st.s = st.s[:l]

	return x
}

func (st *Stack) Push(x any) {
	st.s = append(st.s, x)
}

func (st *Stack) Size() int {
	return len(st.s)
}
