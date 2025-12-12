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

	Kind uint8

	Node struct {
		node node
	}

	// arg | kind | index
	node uint32

	level int8

	Error struct {
		text  string
		index int
	}
)

const (
	indexBits = 16
	kindBits  = 6
	argShift  = indexBits + kindBits
	argBits   = 32 - argShift

	indexMask = 1<<indexBits - 1
	kindMask  = 1<<kindBits - 1
	argMask   = 1<<argBits - 1
)

const (
	_ Kind = iota

	Dot
	Null
	Bool
	Num
	Str

	Pipe
	Comma

	Assign
	PipeAssign

	Or
	And

	Equal
	NotEqual
	Less
	LessEq
	Greater
	GreaterEq

	Add
	Sub

	Mul
	Div
	Mod

	Neg
	Pos

	Arr
	Obj
	Iter
	Index
	Slice

	Var
	Name
	Prop
	Func

	If
	Def

	end      = iota
	Err Kind = 1<<kindBits - (iota - end)
	Comment
)

const (
	levelPipe = iota
	levelComma
	levelAssign
	levelOr
	levelAnd
	levelCmp
	levelAdd
	levelMul
	levelUnary
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

func (p *Parser) parseBinOp(t string, st, lvl int, stopcomma bool) (node, int) {
	//	fmt.Printf("parseBinOp %v\n", st)
	if lvl == levelUnary {
		return p.parseUnOp(t, st)
	}

	l, i := p.parseBinOp(t, st, lvl+1, stopcomma)
	if l.Err() {
		return l, i
	}

	stack := lvl == levelPipe || lvl == levelComma
	var stackop Kind

	reset := len(p.tmp)
	if stack {
		defer func() { p.tmp = p.tmp[:reset] }()

		p.tmp = append(p.tmp, l)
	}

	for {
		i = p.skipSpaces(t, i)
		if i == len(t) {
			break
		}

		op, oplvl, j := p.parseOpBin(t, i)
		if op == Err || stopcomma && op == Comma {
			break
		}
		if oplvl < lvl {
			break
		}

		r, j := p.parseBinOp(t, j, lvl+1, stopcomma)
		if r.Err() {
			return r, j
		}

		if stack {
			p.tmp = append(p.tmp, r)
			stackop = op
		} else {
			l = p.newNodeArg(op, 2, l, r)
		}

		i = j
	}

	ll := len(p.tmp[reset:])

	if !stack || ll == 1 {
		return l, i
	}

	if ll > 256 {
		return p.newErr("pipe overflow", i)
	}

	return p.newNodeArg(stackop, ll, p.tmp[reset:]...), i
}

func (p *Parser) parseUnOp(t string, st int) (node, int) {
	i := p.skipSpaces(t, st)
	if i == len(t) {
		return p.newErr(eof, i)
	}

	var k Kind

	switch t[i] {
	case '+':
		k = Pos
		i++
	case '-':
		k = Neg
		i++
	}

	if k == 0 {
		return p.parseArg(t, i)
	}

	x, i := p.parseUnOp(t, i)
	if x.Err() {
		return x, i
	}

	return p.newNode(k, x), i
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

		return p.newNode(Arr, x), i
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

func (p *Parser) skipComment(t string, i int) int {
	for i < len(t) && t[i] != '\n' {
		if t[i] == '\\' {
			i++
		}

		i++
	}

	return i
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

	const (
		dot = 1 << iota
		prop
		str
		brack
		end
		init
	)

	i = st
	exp := dot | init

	var x node

loop:
	for i < len(t) {
		x = 0

		switch {
		case exp&dot != 0 && t[i] == '.':
			i++
			exp = prop | str | brack | exp&init
		case exp&prop != 0 && skip.IDFirst.Is(t[i]):
			x, i = p.parseName(t, i, Prop)
			exp = dot | brack | end
		case exp&str != 0 && (t[i] == '"' || t[i] == '\''):
			x, i = p.parseString(t, i)
			if !x.Err() {
				x = p.newNode(Index, x)
			}
			exp = dot | brack | end
		case exp&brack != 0 && t[i] == '[':
			x, i = p.parseIndex(t, i)
			exp = dot | brack | end
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
		return p.newNode(Dot), i
	}
	if l == 1 {
		return p.tmp[reset], i
	}
	if l > 256 {
		return p.newErr("pipe overflow", st)
	}

	return p.newNodeArg(Pipe, l, p.tmp[reset:]...), i
}

func (p *Parser) parseIndex(t string, st int) (n node, i int) {
	i = p.skipSpaces(t, st+1)
	if i == len(t) {
		return p.newErr(eof, i)
	}

	if t[i] == ']' {
		i++
		return p.newNode(Iter), i
	}

	var lo, hi node
	var slice bool

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
		slice = true
	}

	if slice && t[i] != ']' {
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

	if !slice {
		return p.newNode(Index, lo), i
	}

	return p.newNode(Slice, lo, hi), i
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

	if l > argMask {
		return p.newErr("object key-value pairs overflow", i)
	}

	return p.newNodeArg(Obj, l, p.tmp[reset:]...), i
}

func (p *Parser) parseObjKey(t string, st int) (node, int) {
	i := p.skipSpaces(t, st)
	if i == len(t) {
		return p.newErr(eof, i)
	}

	if c := t[i]; c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' {
		return p.parseName(t, i, Name)
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
		return p.newNode(Null), end
	case "true":
		return p.newNodeArg(Bool, 1), end
	case "false":
		return p.newNodeArg(Bool, 0), end
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

	l := len(p.tmp[reset:])
	if l > 128 {
		return p.newErr("function arguments overflow", i)
	}

	x := p.newNodeHeader(Func, l)
	p.buf = append(p.buf, node(st), node(end))
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
	if l > argMask {
		return p.newErr("too big if statement", st)
	}

	return p.newNodeArg(If, l, p.tmp[reset:]...), i
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
	if l > argMask {
		return p.newErr("too long string", st)
	}

	return p.newNodeArg(Str, l, node(st)), j
}

func (p *Parser) parseVar(t string, st int) (node, int) {
	s, j := skip.Identifier([]byte(t), st, skip.IDPrefix|'$')
	if s.Err() {
		return p.newErr("bad variable name", st)
	}

	l := j - st - 1
	if l > 128 {
		return p.newErr("too long variable name", st)
	}

	return p.newNodeArg(Var, l, node(st+1)), j
}

func (p *Parser) parseName(t string, st int, kind Kind) (node, int) {
	s, j := skip.Identifier([]byte(t), st, skip.IDUnicode)
	if s.Err() {
		return p.newErr("bad name", st)
	}

	l := j - st
	if l > argMask {
		return p.newErr("too long name", st)
	}

	return p.newNodeArg(kind, l, node(st)), j
}

func (p *Parser) parseNum(t string, st int) (node, int) {
	n, j := skip.Number([]byte(t), st, 0)
	if !n.Ok() {
		return p.newErr("bad number", st)
	}

	l := j - st
	if l > argMask {
		return p.newErr("too long number", st)
	}

	return p.newNodeArg(Num, l, node(st)), j
}

func (p *Parser) parseOpBin(t string, st int) (k Kind, lvl, i int) {
	i = st + 1

	switch t[st] {
	case '|':
		if i < len(t) && t[i] == '|' {
			return Or, levelOr, i + 1
		}
		return Pipe, levelPipe, i
	case ',':
		return Comma, levelComma, i
	case '+':
		return Add, levelAdd, i
	case '-':
		return Sub, levelAdd, i
	case '*':
		return Mul, levelMul, i
	case '/':
		return Div, levelMul, i
	case '%':
		return Mod, levelMul, i
	case '=':
		if i < len(t) && t[i] == '=' {
			return Equal, levelCmp, i + 1
		}
		return Assign, levelAssign, i
	case '!':
		if i < len(t) && t[i] == '=' {
			return NotEqual, levelCmp, i + 1
		}
	case '>':
		if i < len(t) && t[i] == '=' {
			return GreaterEq, levelCmp, i + 1
		}
		return Greater, levelCmp, i
	case '<':
		if i < len(t) && t[i] == '=' {
			return LessEq, levelCmp, i + 1
		}
		return Less, levelCmp, i
	case '&':
		if i < len(t) && t[i] == '&' {
			return And, levelAnd, i + 1
		}
	case 'a':
		if i+1 < len(t) && t[i] == 'n' && t[i+1] == 'd' && (i+2 == len(t) || !skip.IDRest.Is(t[i+2])) {
			return And, levelAnd, i + 2
		}
	case 'o':
		if i < len(t) && t[i] == 'r' && (i+1 == len(t) || !skip.IDRest.Is(t[i+1])) {
			return Or, levelOr, i + 1
		}
	}

	return Err, 0, st
}

func binOpLevel(op Kind) level {
	if op < Pipe || op > Mod {
		return -1
	}

	return []level{
		Pipe:       levelPipe,
		Comma:      levelComma,
		Assign:     levelAssign,
		PipeAssign: levelAssign,
		Or:         levelOr,
		And:        levelAnd,
		Equal:      levelCmp,
		NotEqual:   levelCmp,
		Less:       levelCmp,
		LessEq:     levelCmp,
		Greater:    levelCmp,
		GreaterEq:  levelCmp,
		Add:        levelAdd,
		Sub:        levelAdd,
		Mul:        levelMul,
		Div:        levelMul,
		Mod:        levelMul,
	}[op]
}

func (p *Parser) newErr(t string, i int) (node, int) {
	fmt.Printf("newErr %q %v  from %v\n", t, i, from(1))
	p.err = t

	return p.newNode(Err), i
}

func (p *Parser) newNode(kind Kind, args ...node) node {
	return p.newNodeArg(kind, 0, args...)
}

func (p *Parser) newNodeArg(kind Kind, arg int, args ...node) node {
	x := p.newNodeHeader(kind, arg)
	p.buf = append(p.buf, args...)
	return x
}

func (p *Parser) newNodeHeader(kind Kind, arg int) node {
	id := len(p.buf)
	if id > indexMask {
		panic("index overflow")
	}
	if kind&^kindMask != 0 {
		panic(kind)
	}
	if arg&^(1<<(32-argShift)-1) != 0 {
		panic(arg)
	}

	return node(arg)<<argShift | node(kind)<<indexBits | node(id)
}

func (n node) Index() int { return int(n & indexMask) }
func (n node) Kind() Kind { return Kind(n >> indexBits & kindMask) }
func (n node) Err() bool  { return n.Kind() == Err }
func (n node) Arg() int   { return int(n >> argShift) }

func (n node) GoString() string { return fmt.Sprintf("0x%x", uint(n)) }
func (n node) String() string {
	if n.Arg() == 0 {
		return fmt.Sprintf("%v#%x", n.Kind(), n.Index())
	}

	return fmt.Sprintf("%v#%x(%x)", n.Kind(), n.Index(), n.Arg())
}

func (e *Error) Error() string { return fmt.Sprintf("%v at index %v", e.text, e.index) }

func from(d int) string {
	_, file, line, _ := runtime.Caller(1 + d)

	return fmt.Sprintf("%v:%d", filepath.Base(file), line)
}
