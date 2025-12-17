package jqparser

import (
	"strconv"
	"unsafe"

	"nikand.dev/go/skip"
)

func (n Node) Kind() Kind {
	return Kind(n.node.Kind())
}

func (n Node) Bool() bool {
	return n.node.Arg() != 0
}

func (p *Parser) Bool(n Node) bool {
	return n.Bool()
}

func (p *Parser) Text(n Node) string {
	return p.astext(n.node)
}

func (p *Parser) astext(n node) string {
	switch k := n.Kind(); k {
	case num, str, name, prop, vark, label, brk:
		x := n.Index()
		l := n.Arg()

		st := int(p.nodes[x])

		return p.text[st : st+l]
	default:
		panic(k)
	}
}

func (p *Parser) NumFlags(n Node) skip.Num {
	f, _ := skip.Number([]byte(p.Text(n)), 0, 0)
	return f
}

func (p *Parser) Int(n Node) int {
	return int(p.Int64(n))
}

func (p *Parser) Int64(n Node) int64 {
	x, err := strconv.ParseInt(p.Text(n), 10, 64)
	if err != nil {
		panic(err)
	}

	return x
}

func (p *Parser) Uint64(n Node) uint64 {
	x, err := strconv.ParseUint(p.Text(n), 10, 64)
	if err != nil {
		panic(err)
	}

	return x
}

func (p *Parser) Float64(n Node) float64 {
	x, err := strconv.ParseFloat(p.Text(n), 64)
	if err != nil {
		panic(err)
	}

	return x
}

func (p *Parser) Str(n Node) string {
	switch k := n.node.Kind(); k {
	case name, prop, vark, label, brk:
		return p.Text(n)
	case str:
		s, buf, _, _ := skip.DecodeString([]byte(p.Text(n)), 0, skip.Sqt|skip.Dqt, nil)
		if s.Err() {
			panic(s)
		}

		return unsafe.String(&buf[0], len(buf))
	default:
		panic(n)
	}
}

func (p *Parser) Pipe(n Node) []Node {
	base := n.node.Index()
	l := n.node.Arg()

	args := p.nodes[base : base+l]

	return *(*[]Node)(unsafe.Pointer(&args))
}

func (p *Parser) Comma(n Node) []Node {
	base := n.node.Index()
	l := n.node.Arg()

	args := p.nodes[base : base+l]

	return *(*[]Node)(unsafe.Pointer(&args))
}

func (p *Parser) Bind(n Node) (val, expr Node, bindings []Node) {
	base := n.node.Index()
	l := n.node.Arg()

	args := p.nodes[base+2 : base+2+l]
	bindings = *(*[]Node)(unsafe.Pointer(&args))

	return p.argNode(n, 0), p.argNode(n, 1), bindings
}

func (p *Parser) BinOp(n Node) (op BinOpKind, l, r Node) {
	op = BinOpKind(n.node.Arg())

	return op, p.argNode(n, 0), p.argNode(n, 1)
}

func (p *Parser) UnOp(n Node) (op UnOpKind, x Node) {
	return UnOpKind(n.node.Arg()), p.argNode(n, 0)
}

func (p *Parser) Arr(n Node) Node {
	return p.argNode(n, 0)
}

func (p *Parser) Obj(n Node) []Node {
	base := n.node.Index()
	l := n.node.Arg()

	args := p.nodes[base : base+2*l]

	return *(*[]Node)(unsafe.Pointer(&args))
}

func (p *Parser) Index(n Node) Node {
	return p.argNode(n, 0)
}

func (p *Parser) Slice(n Node) (lo, hi Node) {
	return p.argNode(n, 0), p.argNode(n, 1)
}

func (p *Parser) If(n Node) []Node {
	base := n.node.Index()
	l := n.node.Arg()

	args := p.nodes[base : base+l]

	return *(*[]Node)(unsafe.Pointer(&args))
}

func (p *Parser) FuncCall(n Node) (Node, []Node) {
	base := n.node.Index()
	l := n.node.Arg()

	args := p.nodes[base+1 : base+1+l]

	return p.argNode(n, 0), *(*[]Node)(unsafe.Pointer(&args))
}
