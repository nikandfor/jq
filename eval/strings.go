package eval

import "github.com/nikandfor/errors"

type (
	String []byte
)

func (n String) Parse(p []byte, st int) (b any, i int, err error) {
	if st == len(p) {
		return nil, st, errors.New("string expected")
	}

	i = st

	open := p[i]

	switch open {
	case '"', '`':
		i++
	default:
		return nil, st, errors.New("string expected")
	}

loop:
	for ; i < len(p); i++ {
		r := p[i]

		switch {
		case r == open:
			i++
			break loop
		case r == '\\' && open == '"':
			r = p[i]
			switch r {
			case '\\', 'n', 't':
				panic("support deescaping")
			default:
				return nil, i, errors.New("bad escape esage")
			}
		case r == '\n' && open == '"':
			return nil, i, errors.New("newline in string")
		}
	}

	return String(p[st:i]), i, nil
}

func (s String) String() string {
	if s == nil {
		return "String"
	}

	return string(s)
}
