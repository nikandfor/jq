package jqparser

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"nikand.dev/go/skip"
)

type (
	Parser struct {
		text string
		root node

		nodes []node
		tmp   []node

		err   string
		index int

		prealloc [32]node

		CommaSeparatedArgs bool
		SingleQuoteString  bool
	}

	Kind uint32

	// index | arg | kind1 | kind0
	node uint32

	BinOpKind int16
	UnOpKind  int8

	Node struct {
		node node
	}

	Error struct {
		p *Parser
	}
)

const (
	kind0Bits = 4
	kind1Bits = 4

	arg0Sh  = kind0Bits
	arg1Sh  = kind0Bits + kind1Bits
	indexSh = 16

	kind0Mask = 1<<kind0Bits - 1
	kind1Mask = 1<<(kind0Bits+kind1Bits) - 1

	argPreMask = 1<<indexSh - 1

	maxIndex = 1 << (32 - indexSh - 1)
	maxArg0  = argPreMask >> arg0Sh
	maxArg1  = argPreMask >> arg1Sh
)

const (
	None = Kind(none)
	Dot  = Kind(dot)
	Null = Kind(null)
	Bool = Kind(boolk)

	Num   = Kind(num)
	Str   = Kind(str)
	Name  = Kind(name)
	Prop  = Kind(prop)
	Var   = Kind(vark)
	Label = Kind(label)
	Break = Kind(brk)

	Bind  = Kind(bind)
	Pipe  = Kind(pipe)
	Comma = Kind(comma)

	BinOp = Kind(binop)
	UnOp  = Kind(unop)

	Arr   = Kind(arr)
	Obj   = Kind(obj)
	Iter  = Kind(iter)
	Index = Kind(index)
	Slice = Kind(slice)

	If  = Kind(ifop)
	Try = Kind(try)
	Def = Kind(def)
	//	Let = Kind(let)

	FuncCall = Kind(call)
)

// kond0
const (
	kind1 node = iota
	pipe
	comma
	_

	call
	ifop
	arr
	obj

	num
	str
	name
	prop

	bind
	vark
	label
	brk
)

// kind1
const (
	none = iota<<kind0Bits | kind1
	errnode
	dot
	null

	boolk
	binop
	unop
	noerr

	iter
	index
	slice
	recurse

	try
	def
	_
	_
)

const (
	opPipe BinOpKind = iota << levelSh
	opComma
	Alt
	Assign
	Or
	And
	Equal
	Add
	Mul
	unary

	assreset   = iota
	PipeAssign = Assign + iota - assreset
	AddAssign
	SubAssign
	MulAssign
	DivAssign
	ModAssign
	AltAssign

	eqreset  = iota
	NotEqual = Equal + iota - eqreset
	Less
	LessEq
	Greater
	GreaterEq

	Sub = Add + 1

	Div = Mul + 1
	Mod = Mul + 2

	levelSh = 3
)

const (
	Neg UnOpKind = iota
	Pos
)

const (
	eof = "unexpected end of input"
	sym = "unexpected symbol"
	typ = "unexpected node type"
)

var spaces = skip.NewCharset(" \t\n")

func (p *Parser) Parse(text string) (Node, error) {
	p.Reset()

	if cap(p.nodes) == 0 {
		p.nodes = p.prealloc[:0:24]
	}
	if cap(p.tmp) == 0 {
		p.tmp = p.prealloc[24:24]
	}

	p.text = text

	root, i := p.parseBinOp(text, 0, 0, false)
	if root.Err() {
		return Node{}, Error{p}
	}

	i = p.skipSpaces(text, i)
	if i != len(text) {
		p.err = "trailing data"
		return Node{}, Error{p}
	}

	p.root = root

	return Node{root}, nil
}

func (p *Parser) Reset() {
	p.root = 0
	p.text = ""
	p.nodes = p.nodes[:0]
	p.tmp = p.tmp[:0]
	p.err = ""
	p.index = 0
}

func (p *Parser) Root() Node {
	return Node{p.root}
}

func (p *Parser) parseBinOp(t string, st int, lvl int, stopcomma bool) (node, int) {
	//	fmt.Printf("parseBinOp %v\n", st)
	if lvl == unary.Level() {
		return p.parseUnOp(t, st)
	}

	l, i := p.parseBinOp(t, st, lvl+1, stopcomma)
	if l.Err() {
		return l, i
	}

	chain := lvl == opPipe.Level() || lvl == opComma.Level()

	reset := len(p.tmp)
	if chain {
		defer func() { p.tmp = p.tmp[:reset] }()

		p.tmp = append(p.tmp, l)
	}

	for {
		i = p.skipSpaces(t, i)
		if i == len(t) {
			break
		}

		op, j := p.parseOpBin(t, i)
		if op == -1 || stopcomma && op == opComma {
			break
		}
		if op.Level() < lvl {
			break
		}

		r, j := p.parseBinOp(t, j, lvl+1, stopcomma)
		if r.Err() {
			return r, j
		}

		if chain {
			p.tmp = append(p.tmp, r)
		} else {
			l = p.newBinOp(op, l, r)
		}

		i = j
	}

	ll := len(p.tmp[reset:])

	if !chain || ll == 1 {
		return l, i
	}

	if ll > maxArg0 {
		return p.newErr("chain overflow", i)
	}

	kind := []node{pipe, comma}[lvl]

	return p.newNodeArg0(kind, ll, p.tmp[reset:]...), i
}

func (p *Parser) parseUnOp(t string, st int) (node, int) {
	i := p.skipSpaces(t, st)
	if i == len(t) {
		return p.newErr(eof, i)
	}

	var n node

	switch t[i] {
	case '+':
		n = node(Pos)<<arg1Sh | unop
		i++
	case '-':
		n = node(Neg)<<arg1Sh | unop
		i++
	}

	if n == 0 {
		return p.parseAs(t, i)
	}

	x, i := p.parseUnOp(t, i)
	if x.Err() {
		return x, i
	}

	return p.newNodeArgShifted(n, 0, x), i
}

func (p *Parser) parseAs(t string, st int) (node, int) {
	arg, i := p.parseTerm(t, st)
	if arg.Err() {
		return arg, i
	}

	i = p.skipSpaces(t, i)
	if !(i+2 < len(t) && t[i] == 'a' && t[i+1] == 's' && !skip.IDRest.Is(t[i+2])) {
		return arg, i
	}
	i += 2

	bd, i := p.parseBinding(t, i)
	if bd.Err() {
		return bd, i
	}

	i = p.skipSpaces(t, i)
	if i == len(t) || t[i] != '|' {
		return p.newErr("expected '|' for binding expression", i)
	}
	i++

	expr, i := p.parseBinOp(t, i, 0, true)
	if expr.Err() {
		return expr, i
	}

	return p.newNodeArg0(bind, 1, arg, expr, bd), i
}

func (p *Parser) parseBinding(t string, st int) (n node, i int) {
	i = p.skipSpaces(t, st)

	//	reset := len(p.tmp)
	//	defer func() { p.tmp = p.tmp[:reset] }()

	switch t[i] {
	case '$':
		return p.parseName(t, i+1, vark)
	default:
		return p.newErr("unsupported binding", i)
	}
}

func (p *Parser) parseTerm(t string, st int) (n node, i int) {
	// defer func() { fmt.Printf("parseArg %v -> %v %v  from %v %v\n", st, n, i, from(1), from(2)) }()
	i = p.skipSpaces(t, st)

	switch {
	case t[i] == '(':
		return p.parseParen(t, i)
	case t[i] == '[':
		x, i := p.parseBinOp(t, i+1, 0, false)
		if x.Err() {
			return x, i
		}
		if i == len(t) {
			return p.newErr(eof, i)
		}
		if t[i] != ']' {
			return p.newErr("no closing braket", i)
		}
		i++

		return p.newNodeArg0(arr, 1, x), i
	case t[i] == '{':
		return p.parseObj(t, i)
	case t[i] == '"' || t[i] == '\'' && p.SingleQuoteString:
		return p.parseString(t, i)
	case p.isLiteral(t, i, "null"):
		return p.newNodeArg1NoArgs(null, 0), i + 4
	case p.isLiteral(t, i, "true"):
		return p.newNodeArg1NoArgs(boolk, 1), i + 4
	case p.isLiteral(t, i, "false"):
		return p.newNodeArg1NoArgs(boolk, 0), i + 5
	case p.isLiteral(t, i, "if"):
		return p.parseIf(t, i)
	case t[i] == '.' && i+1 < len(t) && skip.Decimals.Is(t[i+1]) || skip.Decimals.Is(t[i]):
		return p.parseNum(t, i)
	case t[i] == '.', t[i] == '$', skip.WordSymbols.Is(t[i]):
		return p.parseDot(t, i)
	default:
		return p.newErr("unexpected symbol", i)
	}
}

func (p *Parser) isLiteral(t string, i int, word string) bool {
	return strings.HasPrefix(t[i:], word) && (i+len(word) == len(t) || !skip.IDRest.Is(t[i+len(word)]))
}

func (p *Parser) parseDot(t string, st int) (n node, i int) {
	//	defer func() { fmt.Printf("parseDot %v -> %v %v  from %v\n", st, n, i, from(1)) }()

	reset := len(p.tmp)
	defer func() { p.tmp = p.tmp[:reset] }()

	compile := func() node {
		l := len(p.tmp[reset:])
		if l == 0 {
			return dot
		}
		if l == 1 {
			return p.tmp[reset]
		}
		if l > maxArg0 {
			p.err = "members overflow"
			return errnode
		}

		return p.newNodeArg0(pipe, l, p.tmp[reset:]...)
	}

	// .qwe[][].asd.zxc
	// .[][].qwe[]

	type exptyp int

	const (
		expdot exptyp = 1 << iota
		expvar
		expprop
		expname
		expstr
		expbrack
		expcall
		expend
		init
	)

	i = st
	exp := expdot | expvar | expname | init

loop:
	for i < len(t) {
		var x node

		switch {
		case exp&expdot != 0 && t[i] == '.':
			i++
			exp = expprop | expstr | expbrack | expcall | exp&init
		case exp&expvar != 0 && t[i] == '$':
			x, i = p.parseName(t, i+1, vark)
			exp = expdot | expbrack | expcall | expend
		case exp&expname != 0 && skip.IDFirst.Is(t[i]):
			x, i = p.parseName(t, i, name)
			exp = expdot | expbrack | expcall | expend
		case exp&expprop != 0 && skip.IDFirst.Is(t[i]):
			x, i = p.parseName(t, i, prop)
			exp = expdot | expbrack | expcall | expend
		case exp&expstr != 0 && (t[i] == '"' || t[i] == '\'' && p.SingleQuoteString):
			x, i = p.parseString(t, i)
			if !x.Err() {
				x = p.newNodeArg1(index, 0, x)
			}
			exp = expdot | expbrack | expcall | expend
		case exp&expbrack != 0 && t[i] == '[':
			x, i = p.parseIndex(t, i)
			exp = expdot | expbrack | expcall | expend
		case exp&expcall != 0 && t[i] == '(':
			callee := compile()
			if callee.Err() {
				return callee, i
			}

			p.tmp = p.tmp[:reset]

			x, i = p.parseFuncArgs(t, i, callee)

			exp = expdot | expbrack | expcall | expend
		case exp&expend != 0 || exp&init != 0:
			break loop
		default:
			return p.newErr(sym, i)
		}
		if x.Err() {
			return x, i
		}

		if x != 0 {
			p.tmp = append(p.tmp, x)
		}
	}

	return compile(), i
}

func (p *Parser) parseFuncArgs(t string, i int, callee node) (node, int) {
	reset := len(p.tmp)
	defer func() { p.tmp = p.tmp[:reset] }()

	var arg node
	i++

	for {
		i = p.skipSpaces(t, i)

		if i == len(t) {
			return p.newErr(eof, i)
		}
		if t[i] == ')' {
			i++
			break
		}

		arg, i = p.parseBinOp(t, i, 0, p.CommaSeparatedArgs)
		if arg.Err() {
			return arg, i
		}

		p.tmp = append(p.tmp, arg)

		if i < len(t) && (t[i] == ',' || t[i] == ';') {
			i++
		}
	}

	x := p.newNodeArg0(call, len(p.tmp[reset:]))

	p.nodes = append(p.nodes, callee)
	p.nodes = append(p.nodes, p.tmp[reset:]...)

	return x, i
}

func (p *Parser) parseIndex(t string, st int) (n node, i int) {
	i = p.skipSpaces(t, st+1)
	if i == len(t) {
		return p.newErr(eof, i)
	}

	if t[i] == ']' {
		i++
		return p.newNodeArg1NoArgs(iter, 0), i
	}

	var lo, hi node
	var isslice bool

	if t[i] != ':' {
		lo, i = p.parseBinOp(t, i, 0, false)
		if lo.Err() {
			return lo, i
		}
		if i == len(t) {
			return p.newErr(eof, i)
		}
	}

	if t[i] == ':' {
		i++
		isslice = true
	}

	if isslice && t[i] != ']' {
		hi, i = p.parseBinOp(t, i, 0, false)
		if hi.Err() {
			return hi, i
		}
		if i == len(t) {
			return p.newErr(eof, i)
		}
	}

	if t[i] != ']' {
		return p.newErr(sym, i)
	}
	i++

	if !isslice {
		return p.newNodeArg1(index, 0, lo), i
	}

	return p.newNodeArg1(slice, 0, lo, hi), i
}

func (p *Parser) parseObj(t string, st int) (node, int) {
	reset := len(p.tmp)
	defer func() { p.tmp = p.tmp[:reset] }()

	i := st + 1 // '{'
	var k, v node

	for {
		k, i = p.parseObjKey(t, i)
		if k.Err() {
			return k, i
		}
		if i == len(t) {
			return p.newErr(eof, i)
		}

		switch t[i] {
		case ':':
			i++

			v, i = p.parseBinOp(t, i, 0, true)
			if v.Err() {
				return v, i
			}
		case ',', '}':
			if kk := k.Kind(); kk == name {
				v = rekindNodeArg0(prop, k)
			} else if kk == vark {
				v = k
				k = rekindNodeArg0(name, v)
			} else {
				return p.newErr(typ, i)
			}
		default:
			return p.newErr(sym, i)
		}

		p.tmp = append(p.tmp, k, v)

		if t[i] == '}' {
			i++
			break
		}
		if t[i] != ',' {
			return p.newErr(sym, i)
		}
		i++
	}

	l := len(p.tmp[reset:]) / 2

	if l > maxArg0 {
		return p.newErr("object key-value pairs overflow", i)
	}

	return p.newNodeArg0(obj, l, p.tmp[reset:]...), i
}

func (p *Parser) parseObjKey(t string, st int) (node, int) {
	i := p.skipSpaces(t, st)
	if i == len(t) {
		return p.newErr(eof, i)
	}

	if skip.IDFirst.Is(t[i]) {
		return p.parseName(t, i, name)
	}

	if t[i] == '$' {
		return p.parseName(t, i+1, vark)
	}

	if t[i] == '(' {
		return p.parseParen(t, i)
	}

	if t[i] == '"' || t[i] == '\'' && p.SingleQuoteString {
		return p.parseString(t, i)
	}

	return p.newErr(sym, i)
}

func (p *Parser) parseIf(t string, st int) (node, int) {
	reset := len(p.tmp)
	defer func() { p.tmp = p.tmp[:reset] }()

	var cond, then node
	var s skip.ID
	var word string
	i := st + 2 // skip 'if'

	for {
		cond, i = p.parseBinOp(t, i, 0, false)
		if cond.Err() {
			return cond, i
		}

		i = p.skipSpaces(t, i)

		st := i
		s, i = skip.Identifier([]byte(t), i, 0)
		if s.Err() {
			return p.newErr(sym, st)
		}

		if t[st:i] != "then" {
			return p.newErr(sym, st)
		}

		then, i = p.parseBinOp(t, i, 0, false)
		if then.Err() {
			return then, i
		}

		i = p.skipSpaces(t, i)

		st = i
		s, i = skip.Identifier([]byte(t), i, 0)
		if s.Err() {
			return p.newErr(sym, st)
		}

		word = t[st:i]
		p.tmp = append(p.tmp, cond, then)

		switch word {
		case "elif":
			continue
		case "else", "end":
		default:
			return p.newErr(sym, i)
		}

		break
	}

	if word == "else" {
		then, i = p.parseBinOp(t, i, 0, false)
		if then.Err() {
			return then, i
		}

		i = p.skipSpaces(t, i)

		st := i
		s, i = skip.Identifier([]byte(t), i, 0)
		if s.Err() || t[st:i] != "end" {
			return p.newErr(sym, st)
		}

		p.tmp = append(p.tmp, then)
	}

	l := len(p.tmp[reset:])
	if l > maxArg0 {
		return p.newErr("too big if statement", st)
	}

	return p.newNodeArg0(ifop, l, p.tmp[reset:]...), i
}

func (p *Parser) parseParen(t string, st int) (node, int) {
	x, i := p.parseBinOp(t, st+1, 0, false)
	if x.Err() {
		return x, i
	}
	if i == len(t) {
		return p.newErr(eof, i)
	}
	if t[i] != ')' {
		return p.newErr("no closing paren", i)
	}
	i++

	return x, i
}

func (p *Parser) parseString(t string, st int) (node, int) {
	s, _, _, j := skip.String([]byte(t), st, skip.Sqt|skip.Dqt|skip.Unicode)
	if s.Err() {
		return p.newErr("bad string literal", st)
	}

	l := j - st
	if l > maxArg0 {
		return p.newErr("too long string", st)
	}

	return p.newNodeArg0(str, l, node(st)), j
}

func (p *Parser) parseName(t string, st int, kind node) (node, int) {
	s, j := skip.Identifier([]byte(t), st, skip.IDUnicode)
	if s.Err() {
		return p.newErr("bad name", st)
	}

	l := j - st
	if l > maxArg0 {
		return p.newErr("too long name", st)
	}

	return p.newNodeArg0(kind, l, node(st)), j
}

func (p *Parser) parseNum(t string, st int) (node, int) {
	n, j := skip.Number([]byte(t), st, 0)
	if !n.Ok() {
		return p.newErr("bad number", st)
	}

	l := j - st
	if l > maxArg0 {
		return p.newErr("too long number", st)
	}

	return p.newNodeArg0(num, l, node(st)), j
}

func (p *Parser) parseOpBin(t string, st int) (op BinOpKind, i int) {
	i = st + 1

	switch t[st] {
	case '|':
		if i < len(t) && t[i] == '|' {
			return Or, i + 1
		}
		return opPipe, i
	case ',':
		return opComma, i
	case '+':
		return Add, i
	case '-':
		return Sub, i
	case '*':
		return Mul, i
	case '/':
		if i < len(t) && t[i] == '/' {
			return Alt, i + 1
		}
		return Div, i
	case '%':
		return Mod, i
	case '=':
		if i < len(t) && t[i] == '=' {
			return Equal, i + 1
		}
		return Assign, i
	case '!':
		if i < len(t) && t[i] == '=' {
			return NotEqual, i + 1
		}
	case '>':
		if i < len(t) && t[i] == '=' {
			return GreaterEq, i + 1
		}
		return Greater, i
	case '<':
		if i < len(t) && t[i] == '=' {
			return LessEq, i + 1
		}
		return Less, i
	case '&':
		if i < len(t) && t[i] == '&' {
			return And, i + 1
		}
	case 'a':
		if i+1 < len(t) && t[i] == 'n' && t[i+1] == 'd' && (i+2 == len(t) || !skip.IDRest.Is(t[i+2])) {
			return And, i + 2
		}
	case 'o':
		if i < len(t) && t[i] == 'r' && (i+1 == len(t) || !skip.IDRest.Is(t[i+1])) {
			return Or, i + 1
		}
	}

	return -1, st
}

func (op BinOpKind) Level() int {
	return int(op >> levelSh)
}

func (p *Parser) skipSpaces(t string, i int) int {
	for {
		i = spaces.Skip([]byte(t), i)
		if i == len(t) || t[i] != '#' {
			break
		}

		i = p.skipComment(t, i)
	}

	return i
}

func (p *Parser) skipComment(t string, i int) int {
	for i < len(t) && t[i] != '\n' {
		if t[i] == '\\' {
			i++
		}

		i++
	}

	return i
}

func (p *Parser) newErr(t string, i int) (node, int) {
	//	fmt.Printf("newErr %q %v  from %v\n", t, i, from(1))
	p.err = t

	return errnode, i
}

func (p *Parser) newBinOp(op BinOpKind, l, r node) node {
	return p.newNodeArg1(binop, int(op), l, r)
}

func (p *Parser) newNodeArg0(kind node, arg0 int, args ...node) node {
	if arg0 > maxArg0 {
		panic(arg0)
	}

	return p.newNodeArgShifted(kind, arg0<<arg0Sh, args...)
}

func (p *Parser) newNodeArg1(kind node, arg1 int, args ...node) node {
	if arg1 > maxArg1 {
		panic(arg1)
	}

	return p.newNodeArgShifted(kind, arg1<<arg1Sh, args...)
}

func (p *Parser) newNodeArg1NoArgs(kind node, arg1 int) node {
	if arg1 > maxArg1 {
		panic(arg1)
	}

	return node(arg1)<<arg1Sh | kind
}

func (p *Parser) newNodeArgShifted(kind node, arg int, args ...node) node {
	idx := len(p.nodes)
	if idx > maxIndex {
		panic("index overflow")
	}

	//	fmt.Printf("newNode  %#v  %v  %v  from %v %v\n", kind, arg, args, from(2), from(3))

	x := makeNode(kind, arg, idx)

	p.nodes = append(p.nodes, args...)

	return x
}

func makeNode(kind node, arg, idx int) node {
	return node(idx)<<indexSh | node(arg) | kind
}

func rekindNodeArg0(kind, n node) node {
	return n&^kind0Mask | kind
}

func (n node) Kind0() node   { return n & kind0Mask }
func (n node) Kind1() node   { return n & kind1Mask }
func (n node) IsKind1() bool { return n.Kind0() == kind1 }
func (n node) Kind() node {
	if n.IsKind1() {
		return n & kind1Mask
	}
	return n & kind0Mask
}
func (n node) Index() int { return int(n >> indexSh) }
func (n node) Err() bool  { return n == errnode }
func (n node) Arg() int {
	if n.IsKind1() {
		return int(n & argPreMask >> arg1Sh)
	}

	return int(n & argPreMask >> arg0Sh)
}

func (n node) body() node {
	if n.IsKind1() {
		return n &^ kind1Mask
	}

	return n &^ kind0Mask
}

func (n node) toOp() BinOpKind {
	switch n.Kind() {
	case pipe:
		return opPipe
	case comma:
		return opComma
	default:
		panic(n)
	}
}

func (p *Parser) argNode(n Node, i int) Node { return Node{p.nodes[n.node.Index()+i]} }
func (p *Parser) arg(n node, i int) node     { return p.nodes[n.Index()+i] }
func (p *Parser) argInt(n node, i int) int   { return int(p.arg(n, i)) }

func (n node) GoString() string {
	return fmt.Sprintf("0x%x_%x_%x", n.Index(), n.Arg(), int(n.Kind()))
}

//func (n node) String() string {
//	return fmt.Sprintf("%v#%x(%x)", Node{n}.Kind(), n.Index(), n.Arg())
//}

func (e Error) Error() string { return fmt.Sprintf("%v at index %v", e.p.text, e.p.index) }

func from(d int) string {
	_, file, line, _ := runtime.Caller(1 + d)

	return fmt.Sprintf("%v:%d", filepath.Base(file), line)
}
