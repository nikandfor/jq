package jq

func NewQuery(fs ...any) *Pipe {
	return NewPipe(query(fs)...)
}

func NewMulti(fs ...any) *Comma {
	return NewComma(query(fs)...)
}

func query(p []any) []Filter {
	q := make([]Filter, len(p))

	for i := range p {
		switch x := p[i].(type) {
		case int:
			q[i] = Index(x)
		case string:
			q[i] = Key(x)
		case Iter:
			q[i] = &x
		case Filter:
			q[i] = x
		default:
			panic(p[i])
		}
	}

	return q
}
