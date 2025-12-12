package jqparser

func (p *Parser) Text(n Node) string {
	return p.astext(n.node)
}

func (p *Parser) astext(n node) string {
	x := n.Index()

	switch n.Kind() {
	case Num, Str, Name, Prop, Var:
	case Func:
		st := int(p.buf[x])
		end := int(p.buf[x+1])

		return p.text[st:end]
	default:
		panic(n.Kind())
	}

	l := n.Arg()

	st := int(p.buf[x])

	return p.text[st : st+l]
}
