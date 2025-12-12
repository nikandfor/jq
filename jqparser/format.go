package jqparser

import (
	"fmt"
	"strings"
	"unsafe"

	"nikand.dev/go/skip"
)

func (p *Parser) Format(n Node) string {
	b := p.AppendFormat(nil, n)

	return unsafe.String(unsafe.SliceData(b), len(b))
}

func (p *Parser) AppendFormat(b []byte, n Node) []byte {
	return p.appendFormat(b, n.node, -1)
}

func (p *Parser) appendFormat(b []byte, n node, lvl level) []byte {
	op := func(n node, op string, level level) []byte {
		x := n.Index()

		if lvl > level {
			b = append(b, '(')
		}

		b = p.appendFormat(b, p.buf[x], level)
		b = append(b, op...)

		rlevel := binOpLevel(p.buf[x+1].Kind())
		if rlevel == level {
			b = append(b, '(')
		}

		b = p.appendFormat(b, p.buf[x+1], level)

		if rlevel == level {
			b = append(b, ')')
		}

		if lvl > level {
			b = append(b, ')')
		}

		return b
	}

	switch n.Kind() {
	case 0:
		return append(b, "<nil>"...)
	case Err:
		return append(b, "<error>"...)
	case Dot:
		return append(b, '.')
	case Null:
		return append(b, "null"...)
	case Bool:
		if n.Arg() != 0 {
			return append(b, "true"...)
		}

		return append(b, "false"...)
	case Num:
		return append(b, p.astext(n)...)
	case Str:
		return append(b, p.astext(n)...)
	case Pipe:
		return op(n, " | ", levelPipe)
	case Comma:
		return op(n, ", ", levelComma)
	case Assign:
		return op(n, " = ", levelAssign)
	case PipeAssign:
		return op(n, " |= ", levelAssign)
	case Or:
		return op(n, " or ", levelOr)
	case And:
		return op(n, " and ", levelAnd)
	case Equal:
		return op(n, " == ", levelCmp)
	case NotEqual:
		return op(n, " != ", levelCmp)
	case Less:
		return op(n, " < ", levelCmp)
	case LessEq:
		return op(n, " <= ", levelCmp)
	case Greater:
		return op(n, " > ", levelCmp)
	case GreaterEq:
		return op(n, " >= ", levelCmp)
	case Add:
		return op(n, " + ", levelAdd)
	case Sub:
		return op(n, " - ", levelAdd)
	case Mul:
		return op(n, " * ", levelMul)
	case Div:
		return op(n, " / ", levelMul)
	case Mod:
		return op(n, " % ", levelMul)
	case Pos:
		x := n.Index()
		b = append(b, '+')
		return p.appendFormat(b, p.buf[x], levelUnary)
	case Neg:
		x := n.Index()
		b = append(b, '-')
		return p.appendFormat(b, p.buf[x], levelUnary)
	case Arr:
		x := n.Index()
		b = append(b, '[')
		b = p.appendFormat(b, p.buf[x], -1)
		b = append(b, ']')
		return b
	case Obj:
		x := n.Index()
		l := n.Arg()
		b = append(b, '{')
		for i := range l {
			if i != 0 {
				b = append(b, ',', ' ')
			}

			k := p.buf[x+2*i]
			v := p.buf[x+2*i+1]

			kind := k.Kind()
			q := kind == Name || kind == Var || kind == Str

			if !q {
				b = append(b, '(')
			}
			b = p.appendFormat(b, k, -1)
			if !q {
				b = append(b, ')')
			}

			if k == v {
				continue
			}

			b = append(b, ':', ' ')
			b = p.appendFormat(b, v, -1)
		}
		b = append(b, '}')
		return b
	case Iter:
		return append(b, ".[]"...)
	case Index:
		x := n.Index()
		b = append(b, '.', '[')
		b = p.appendFormat(b, p.buf[x], -1)
		return append(b, ']')
	case Slice:
		x := n.Index()
		b = append(b, '.', '[')

		if p.buf[x] != 0 {
			b = p.appendFormat(b, p.buf[x], -1)
		}

		b = append(b, ':')

		if p.buf[x+1] != 0 {
			b = p.appendFormat(b, p.buf[x+1], -1)
		}

		return append(b, ']')
	case Var:
		b = append(b, '$')
		return append(b, p.astext(n)...)
	case Name:
		return append(b, p.astext(n)...)
	case Prop:
		b = append(b, '.')
		return append(b, p.astext(n)...)
	case Func:
		x := n.Index()
		l := n.Arg()

		b = append(b, p.astext(n)...)

		if l == 0 {
			return b
		}

		b = append(b, '(')

		for i := range l {
			if i != 0 {
				b = append(b, ',', ' ')
			}

			a := p.buf[x+2+i]
			k := a.Kind()
			par := k == Pipe || k == Comma

			if par {
				b = append(b, '(')
			}

			b = p.appendFormat(b, a, -1)

			if par {
				b = append(b, ')')
			}
		}

		b = append(b, ')')

		return b
	case If:
		x := n.Index()
		l := n.Arg()
		i := 0

		for i+1 < l {
			if i == 0 {
				b = append(b, "if "...)
			} else {
				b = append(b, " elif "...)
			}

			b = p.appendFormat(b, p.buf[x+i], -1)

			b = append(b, " then "...)

			b = p.appendFormat(b, p.buf[x+i+1], -1)

			i += 2
		}

		if i < l {
			b = append(b, " else "...)
			b = p.appendFormat(b, p.buf[x+i], -1)
		}

		b = append(b, " end"...)

		return b
	default:
		panic(n)
	}
}

func (p *Parser) Where(err error) string {
	e, ok := err.(*Error)
	if !ok {
		return ""
	}

	line := 1
	linest := 0
	i := 0

	for i < e.index {
		if p.text[i] == '\n' {
			line++
			linest = i
		}

		i++
	}

	end := skip.NewCharset("\n").SkipUntil([]byte(p.text), i)

	var b strings.Builder

	fmt.Fprintf(&b, "@%s@  (line %d, off %d, col %d)\n", p.text[linest:end], line, linest, i-linest)

	return b.String()
}

func (n Node) Kind() Kind       { return n.node.Kind() }
func (n Node) String() string   { return fmt.Sprintf("%v#%d", n.node.Kind().String(), n.node.Index()) }
func (n Node) GoString() string { return fmt.Sprintf("0x%x", int(n.node)) }

func (k Kind) String() string {
	switch k {
	case 0:
		return "none"
	case Err:
		return "error"
	case Dot:
		return "dot"
	case Null:
		return "null"
	case Bool:
		return "bool"
	case Num:
		return "num"
	case Str:
		return "str"
	case Pipe:
		return "pipe"
	case Comma:
		return "comma"
	case Assign:
		return "assign"
	case PipeAssign:
		return "pipe_assign"
	case Or:
		return "or"
	case And:
		return "and"
	case Equal:
		return "equal"
	case NotEqual:
		return "not_equal"
	case Less:
		return "less"
	case LessEq:
		return "less_eq"
	case Greater:
		return "greater"
	case GreaterEq:
		return "greater_eq"
	case Add:
		return "add"
	case Sub:
		return "sub"
	case Mul:
		return "mul"
	case Div:
		return "div"
	case Mod:
		return "mod"
	case Neg:
		return "neg"
	case Pos:
		return "pos"
	case Arr:
		return "arr"
	case Obj:
		return "obj"
	case Iter:
		return "iter"
	case Var:
		return "var"
	case Name:
		return "name"
	case Prop:
		return "prop"
	case Index:
		return "index"
	case Slice:
		return "slice"
	case Func:
		return "func"
	case If:
		return "if"
	case Def:
		return "def"
	default:
		return fmt.Sprintf("0x%x", int(k))
	}
}
