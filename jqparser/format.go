package jqparser

import (
	"fmt"
	"strings"
	"unsafe"

	"nikand.dev/go/skip"
)

func (p *Parser) Format(n Node) string {
	switch n.node.Kind() {
	case num, str:
		return p.astext(n.node)
	}

	b := p.AppendFormat(nil, n)

	return unsafe.String(unsafe.SliceData(b), len(b))
}

func (p *Parser) AppendFormat(b []byte, n Node) []byte {
	return p.appendFormat(b, n.node, -1, false)
}

func (p *Parser) appendFormat(b []byte, n node, parlevel level, train bool) []byte {
	chain := func(n node) []byte {
		x := n.Index()
		oplevel := levelPipe
		sign := " | "
		if n.Kind() == comma {
			oplevel = levelComma
			sign = ", "
		}

		if parlevel > oplevel {
			b = append(b, '(')
		}

		b = p.appendFormat(b, p.buf[x], oplevel, false)

		cnt := n.Arg()

		for i := 1; i < cnt; i++ {
			r := p.buf[x+i]

			if k := r.Kind(); oplevel == levelPipe && (k == prop || k == iter || k == index || k == slice) {
				b = p.appendFormat(b, r, oplevel, true)
				continue
			}

			b = append(b, sign...)

			rlevel := binOpLevel(BinOpKind(r.Arg()))
			if rlevel == oplevel {
				b = append(b, '(')
			}

			b = p.appendFormat(b, r, oplevel, false)

			if rlevel == oplevel {
				b = append(b, ')')
			}
		}

		if parlevel > oplevel {
			b = append(b, ')')
		}

		return b
	}

	op := func(n node) []byte {
		l, r := p.Arg(n, 0), p.Arg(n, 1)
		op := BinOpKind(n.Arg())
		sign := op.String()
		oplevel := binOpLevel(op)

		if parlevel > oplevel {
			b = append(b, '(')
		}

		b = p.appendFormat(b, l, oplevel, false)

		b = append(b, ' ')
		b = append(b, sign...)
		b = append(b, ' ')

		rlevel := binOpLevel(BinOpKind(r.Arg()))
		if rlevel == oplevel {
			b = append(b, '(')
		}

		b = p.appendFormat(b, r, oplevel, false)

		if rlevel == oplevel {
			b = append(b, ')')
		}

		if parlevel > oplevel {
			b = append(b, ')')
		}

		return b
	}

	switch n.Kind() {
	case none:
		return append(b, "<nil>"...)
	case errk:
		return append(b, "<error>"...)
	case dot:
		return append(b, '.')
	case null:
		return append(b, "null"...)
	case boolk:
		if n.Arg() != 0 {
			return append(b, "true"...)
		}

		return append(b, "false"...)
	case num, str, name:
		return append(b, p.astext(n)...)
	case vark:
		b = append(b, '$')
		return append(b, p.astext(n)...)
	case prop:
		b = append(b, '.')
		return append(b, p.astext(n)...)
	case pipe, comma:
		return chain(n)
	case binop:
		return op(n)
	case unop:
		switch op := UnOpKind(n.Arg()); op {
		case Pos:
			x := n.Index()
			b = append(b, '+')
			return p.appendFormat(b, p.buf[x], levelUnary, false)
		case Neg:
			x := n.Index()
			b = append(b, '-')
			return p.appendFormat(b, p.buf[x], levelUnary, false)
		default:
			panic(op)
		}
	case arr:
		x := n.Index()
		b = append(b, '[')
		b = p.appendFormat(b, p.buf[x], -1, false)
		b = append(b, ']')
		return b
	case obj:
		x := n.Index()
		l := n.Arg()
		b = append(b, '{')
		for i := range l {
			if i != 0 {
				b = append(b, ',', ' ')
			}

			k := p.buf[x+2*i]
			v := p.buf[x+2*i+1]

			kk := k.Kind()
			par := !(kk == name || kk == vark || kk == str)

			if par {
				b = append(b, '(')
			}
			b = p.appendFormat(b, k, -1, false)
			if par {
				b = append(b, ')')
			}

			if k == v {
				continue
			}

			b = append(b, ':', ' ')
			b = p.appendFormat(b, v, -1, false)
		}
		b = append(b, '}')
		return b
	case iter:
		if !train {
			b = append(b, '.')
		}

		return append(b, "[]"...)
	case index:
		if !train {
			b = append(b, '.')
		}

		x := n.Index()
		b = append(b, '[')
		b = p.appendFormat(b, p.buf[x], -1, false)
		return append(b, ']')
	case slice:
		if !train {
			b = append(b, '.')
		}

		x := n.Index()
		b = append(b, '[')

		if p.buf[x] != 0 {
			b = p.appendFormat(b, p.buf[x], -1, false)
		}

		b = append(b, ':')

		if p.buf[x+1] != 0 {
			b = p.appendFormat(b, p.buf[x+1], -1, false)
		}

		return append(b, ']')
	case fun:
		x := n.Index()
		l := p.ArgInt(n, 1)

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
			par := k == pipe || k == comma

			if par {
				b = append(b, '(')
			}

			b = p.appendFormat(b, a, -1, false)

			if par {
				b = append(b, ')')
			}
		}

		b = append(b, ')')

		return b
	case ifop:
		x := n.Index()
		l := n.Arg()
		i := 0

		for i+1 < l {
			if i == 0 {
				b = append(b, "if "...)
			} else {
				b = append(b, " elif "...)
			}

			b = p.appendFormat(b, p.buf[x+i], -1, false)

			b = append(b, " then "...)

			b = p.appendFormat(b, p.buf[x+i+1], -1, false)

			i += 2
		}

		if i < l {
			b = append(b, " else "...)
			b = p.appendFormat(b, p.buf[x+i], -1, false)
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

	fmt.Fprintf(&b, "@%s@  (line %d, col %d)\n", p.text[linest:end], line, i-linest)

	return b.String()
}

func (n Node) String() string   { return fmt.Sprintf("%v#%d", n.node.Kind().String(), n.node.Index()) }
func (n Node) GoString() string { return fmt.Sprintf("0x%x", int(n.node)) }

func (k Kind) String() string {
	switch k {
	case None:
		return "none"
	case Dot:
		return "dot"
	case Null:
		return "null"
	case Bool:
		return "bool"
	case Name:
		return "name"
	case Num:
		return "num"
	case Str:
		return "str"
	case Bind:
		return "bind"
	case Var:
		return "var"
	case Label:
		return "label"
	case Break:
		return "break"
	case Pipe:
		return "pipe"
	case Comma:
		return "comma"
	case BinOp:
		return "binop"
	case UnOp:
		return "unop"
	case Arr:
		return "arr"
	case Obj:
		return "obj"
	case Iter:
		return "iter"
	case Index:
		return "index"
	case Slice:
		return "slice"
	case If:
		return "if"
	case Try:
		return "try"
	case Def:
		return "def"
	case Func:
		return "func"
	default:
		return fmt.Sprintf("0x%x", int(k))
	}
}

func (p BinOpKind) String() string {
	opsign := []string{
		Alt:       "//",
		Or:        "or",
		And:       "and",
		Equal:     "==",
		NotEqual:  "!=",
		Less:      "<",
		LessEq:    "<=",
		Greater:   ">",
		GreaterEq: ">=",
		Add:       "+",
		Sub:       "-",
		Mul:       "*",
		Div:       "/",
		Mod:       "%",
	}[p&0xf]

	if p&Assign != 0 {
		return opsign + "="
	}

	return opsign
}
