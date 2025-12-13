package jqparser

func (p *Parser) Text(n Node) string {
	return p.astext(n.node)
}

func (p *Parser) astext(n node) string {
	switch k := n.Kind(); k {
	case num, str, name, prop, vark, bind, label, brk, fun:
		x := n.Index()
		l := n.Arg()

		st := int(p.buf[x])

		return p.text[st : st+l]
	default:
		panic(k)
	}
}
