package jqparser

import (
	"fmt"
	"path/filepath"
	"runtime"

	"nikand.dev/go/skip"
)

type (
	Parser struct {
		root node
		text string

		buf []node
		tmp []node

		err string
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
		text  string
		index int
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

	maxID   = 1 << (32 - indexSh - 1)
	maxArg0 = argPreMask >> arg0Sh
	maxArg1 = argPreMask >> arg1Sh
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
	Bind  = Kind(bind)
	Var   = Kind(vark)
	Label = Kind(label)
	Break = Kind(brk)

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

	Func = Kind(fun)
)

// kond0
const (
	kind1 node = iota
	pipe
	comma
	_

	fun
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
	errk
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
)

var spaces = skip.NewCharset(" \t\n")

func (p *Parser) Parse(text string) (Node, error) {
	p.Reset()

	p.text = text

	root, i := p.parseBinOp(text, 0, 0, false)
	if root.Err() {
		return Node{}, &Error{text: p.err, index: i}
	}

	i = p.skipSpaces(text, i)
	if i != len(text) {
		return Node{}, &Error{text: "trailing data", index: i}
	}

	p.root = root

	return p.Root(), nil
}

func (p *Parser) Reset() {
	p.root = 0
	p.text = ""
	p.buf = p.buf[:0]
	p.tmp = p.tmp[:0]
	p.err = ""
}

func (p *Parser) Root() Node {
	return Node{p.root}
}

func (p *Parser) parseExpr(t string, st int, stopcomma bool) (node, int) {
	x, i := p.parseBinOp(t, st, 0, stopcomma)
	if x.Err() {
		return x, i
	}

	return x, i
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
		return p.parseArg(t, i)
	}

	x, i := p.parseUnOp(t, i)
	if x.Err() {
		return x, i
	}

	return p.newNodeArgShifted(n, 0, x), i
}

func (p *Parser) parseArg(t string, st int) (n node, i int) {
	//	defer func() { fmt.Printf("parseArg %v -> %v %v  from %v %v\n", st, n, i, from(1), from(2)) }()
	i = p.skipSpaces(t, st)

	switch {
	case t[i] == '.':
		return p.parseDot(t, i)
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
	case t[i] == '$':
		return p.parseVar(t, i)
	case t[i] == '"' || t[i] == '\'':
		return p.parseString(t, i)
	case skip.Decimals.Is(t[i]):
		return p.parseNum(t, i)
	case skip.WordSymbols.Is(t[i]):
		return p.parseWord(t, i)
	default:
		return p.newErr("unexpected symbol", i)
	}
}

func (p *Parser) parseDot(t string, st int) (n node, i int) {
	//	defer func() { fmt.Printf("parseDot %v -> %v %v  from %v\n", st, n, i, from(1)) }()

	if st+1 < len(t) && skip.Decimals.Is(t[st+1]) {
		return p.parseNum(t, st)
	}

	reset := len(p.tmp)
	defer func() { p.tmp = p.tmp[:reset] }()

	// .qwe[][].asd.zxc
	// .[][].qwe[]

	type exptyp int

	const (
		expdot exptyp = 1 << iota
		expprop
		str
		brack
		end
		init
	)

	i = st
	exp := expdot | init

	var x node

loop:
	for i < len(t) {
		x = 0

		switch {
		case exp&expdot != 0 && t[i] == '.':
			i++
			exp = expprop | str | brack | exp&init
		case exp&expprop != 0 && skip.IDFirst.Is(t[i]):
			x, i = p.parseName(t, i, prop)
			exp = expdot | brack | end
		case exp&str != 0 && (t[i] == '"' || t[i] == '\''):
			x, i = p.parseString(t, i)
			if !x.Err() {
				x = p.newNodeArg1(index, 0, x)
			}
			exp = expdot | brack | end
		case exp&brack != 0 && t[i] == '[':
			x, i = p.parseIndex(t, i)
			exp = expdot | brack | end
		case exp&end != 0 || exp&init != 0:
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

	l := len(p.tmp[reset:])
	if l == 0 {
		return p.newNodeNoArgs(dot, 0), i
	}
	if l == 1 {
		return p.tmp[reset], i
	}
	if l > maxArg0 {
		return p.newErr("pipe overflow", st)
	}

	return p.newNodeArg0(pipe, l, p.tmp[reset:]...), i
}

func (p *Parser) parseIndex(t string, st int) (n node, i int) {
	i = p.skipSpaces(t, st+1)
	if i == len(t) {
		return p.newErr(eof, i)
	}

	if t[i] == ']' {
		i++
		return p.newNodeNoArgs(iter, 0), i
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

		if t[i] == ':' {
			i++

			v, i = p.parseBinOp(t, i, 0, true)
			if v.Err() {
				return v, i
			}
		} else if t[i] == ',' || t[i] == '}' {
			v = k
		} else {
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

	if c := t[i]; c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' {
		return p.parseName(t, i, name)
	}

	if t[i] == '$' {
		return p.parseVar(t, i)
	}

	if t[i] == '(' {
		return p.parseParen(t, i)
	}

	if t[i] == '"' || t[i] == '\'' {
		return p.parseString(t, i)
	}

	return p.newErr(sym, i)
}

func (p *Parser) parseWord(t string, st int) (node, int) {
	s, end := skip.Identifier([]byte(t), st, 0)
	if s.Err() {
		return p.newErr("bad filter name", st)
	}

	switch t[st:end] {
	case "null":
		return p.newNodeNoArgs(null, 0), end
	case "true":
		return p.newNodeNoArgs(boolk, 1), end
	case "false":
		return p.newNodeNoArgs(boolk, 0), end
	case "if":
		return p.parseIf(t, st)
	}

	reset := len(p.tmp)
	defer func() { p.tmp = p.tmp[:reset] }()

	i := p.skipSpaces(t, end)
	if i < len(t) && t[i] == '(' {
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

			arg, i = p.parseBinOp(t, i, 0, true)
			if arg.Err() {
				return arg, i
			}

			p.tmp = append(p.tmp, arg)

			if i < len(t) && (t[i] == ',' || t[i] == ';') {
				i++
			}
		}
	}

	//	l := len(p.tmp[reset:])
	l := end - st
	if l > maxArg0 {
		return p.newErr("function name overflow", i)
	}

	x := p.newNodeArg0(fun, l)

	p.buf = append(p.buf, node(st), node(len(p.tmp[reset:])))
	p.buf = append(p.buf, p.tmp[reset:]...)

	return x, i
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

func (p *Parser) parseVar(t string, st int) (node, int) {
	s, j := skip.Identifier([]byte(t), st, skip.IDPrefix|'$')
	if s.Err() {
		return p.newErr("bad variable name", st)
	}

	l := j - st - 1
	if l > maxArg0 {
		return p.newErr("too long variable name", st)
	}

	return p.newNodeArg0(vark, l, node(st+1)), j
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
	fmt.Printf("newErr %q %v  from %v\n", t, i, from(1))
	p.err = t

	return errk, i
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

func (p *Parser) newNodeNoArgs(kind node, arg1 int) node {
	if arg1 > maxArg1 {
		panic(arg1)
	}

	return node(arg1)<<arg1Sh | kind
}

func (p *Parser) newNodeArgShifted(kind node, arg int, args ...node) node {
	id := len(p.buf)
	if id > maxID {
		panic("index overflow")
	}

	x := node(id)<<indexSh | node(arg) | kind

	p.buf = append(p.buf, args...)

	return x
}

func (n Node) Kind() Kind {
	return Kind(n.node.Kind())
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
func (n node) Err() bool  { return n == errk }
func (n node) Arg() int {
	if n.IsKind1() {
		return int(n & argPreMask >> arg1Sh)
	}

	return int(n & argPreMask >> arg0Sh)
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

func (p *Parser) Arg(n node, i int) node   { return p.buf[n.Index()+i] }
func (p *Parser) ArgInt(n node, i int) int { return int(p.Arg(n, i)) }

func (n node) GoString() string {
	return fmt.Sprintf("0x%x_%x_%x", n.Index(), n.Arg(), int(n.Kind()))
}

func (n node) String() string {
	return fmt.Sprintf("%v#%x(%x)", Node{n}.Kind(), n.Index(), n.Arg())
}

func (e *Error) Error() string { return fmt.Sprintf("%v at index %v", e.text, e.index) }

func from(d int) string {
	_, file, line, _ := runtime.Caller(1 + d)

	return fmt.Sprintf("%v:%d", filepath.Base(file), line)
}
