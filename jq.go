package jq

import (
	"errors"
	"fmt"
	"strings"

	"nikand.dev/go/cbor"
)

type (
	Filter interface {
		ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error)
	}

	FilterFunc func(b *Buffer, off Off, next bool) (Off, bool, error)

	FilterPath interface {
		ApplyToGetPath(b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error)
	}

	NodePath []NodePathSeg

	NodePathSeg struct {
		Off   Off
		Index int
		Key   Off
	}

	Off   int
	Dot   struct{}
	Empty struct{}
	Halt  struct {
		Err error
	}
	Literal []byte

	TypeError int
)

const (
	_ Off = -iota
	None
	Null
	True
	False
	Zero
	One
)

// ErrType = errors.New("type error")
var ErrHalt = errors.New("halted")

func ApplyGetPath(f Filter, b *Buffer, off Off, base NodePath, next bool) (res Off, path NodePath, more bool, err error) {
	fp, ok := f.(FilterPath)
	if ok {
		return fp.ApplyToGetPath(b, off, base, next)
	}

	path = base
	res, more, err = f.ApplyTo(b, off, next)
	return
}

func (f FilterFunc) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	return f(b, off, next)
}

func (f Off) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	if next {
		return off, false, nil
	}

	return f, false, nil
}

func (f Dot) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	if next {
		return None, false, nil
	}

	return off, false, nil
}

func (f Empty) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	return None, false, nil
}

func (f Halt) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	err := f.Err
	if err == nil {
		err = ErrHalt
	}

	return off, false, err
}

func NewLiteral(x any) Literal {
	var e Encoder

	switch x := x.(type) {
	case string:
		return Literal(e.AppendString(nil, x))
	case int:
		return Literal(e.AppendInt(nil, x))
	case int64:
		return Literal(e.AppendInt64(nil, x))
	case bool:
		return Literal(e.AppendBool(nil, x))
	}

	panic(x)
}

func (f Literal) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	if next {
		return None, false, nil
	}

	return b.Writer().Raw(f), false, nil
}

func (f Off) String() string { return fmt.Sprintf("%x", int(f)) }
func (f Off) Format(s fmt.State, v rune) {
	if v == 'v' {
		v = 'x'
	}

	_, _ = fmt.Fprintf(s, "%"+string(v), int64(f))
}

func (f Dot) String() string   { return "." }
func (f Empty) String() string { return "empty" }

func (f Literal) String() string {
	var d Decoder

	tag := d.TagOnly(f, 0)

	switch tag {
	case cbor.Int, cbor.Neg:
		v, i := d.Unsigned(f, 0)
		if i == len(f) {
			minus := ""
			if tag == cbor.Neg {
				minus = "-"
			}

			return fmt.Sprintf("%s%d", minus, v)
		}
	case cbor.String:
		v, i := d.Bytes(f, 0)
		if i == len(f) {
			return fmt.Sprintf("%q", v)
		}
	case cbor.Simple:
		_, sub, i := d.CBOR.Tag(f, 0)
		if i != len(f) {
			break
		}

		switch sub {
		case cbor.False:
			return "false"
		case cbor.True:
			return "true"
		case cbor.Null:
			return "null"
		case cbor.Float8, cbor.Float16, cbor.Float32, cbor.Float64:
			v, _ := d.Float(f, 0)
			return fmt.Sprintf("%v", v)
		}
	}

	return fmt.Sprintf("Literal(%#x)", []byte(f))
}

func appendHex(b, a []byte) []byte {
	const digits = "0123456789abcdef"

	for i, c := range a {
		if i != 0 {
			b = append(b, ' ')
		}

		b = append(b, digits[c>>4], digits[c&0xf])
	}

	return b
}

func (p NodePath) Format(s fmt.State, v rune) {
	if len(p) == 0 {
		_, _ = s.Write([]byte{'/'})
		return
	}

	for i, ps := range p {
		if i != 0 {
			_, _ = s.Write([]byte{'/'})
		}

		_, _ = fmt.Fprintf(s, "%"+string(v), ps)
	}
}

func (p NodePath) String() string {
	if len(p) == 0 {
		return "/"
	}

	var b strings.Builder

	for i, ps := range p {
		if i != 0 {
			_ = b.WriteByte('/')
		}

		_, _ = fmt.Fprintf(&b, "%x", ps)
	}

	return b.String()
}

func (p NodePathSeg) Format(s fmt.State, v rune) {
	p.Off.Format(s, v)
	_, _ = fmt.Fprintf(s, ":%"+string(v), p.Index)

	if p.Key != None {
		fmt.Fprintf(s, "(%v)", p.Key)
	}
}

func (p NodePathSeg) String() string {
	return fmt.Sprintf("%v:%x", p.Off, p.Index)
}

func NewTypeError(got byte, wanted ...byte) TypeError {
	var e TypeError

	e |= TypeError(got)

	for _, t := range wanted {
		e |= 1 << (8 + t>>5)
	}

	return e
}

func (e TypeError) Error() string {
	var b strings.Builder

	tag := byte(e)

	_, _ = fmt.Fprintf(&b, "type error: %s (%x)", tagString(tag), tag)

	return b.String()
}

func tagString(tag byte) string {
	return tag2str[tag>>5]
}

var tag2str = []string{
	cbor.Int >> 5:     "Int",
	cbor.Neg >> 5:     "Neg",
	cbor.Bytes >> 5:   "Bytes",
	cbor.String >> 5:  "String",
	cbor.Array >> 5:   "Array",
	cbor.Map >> 5:     "Map",
	cbor.Simple >> 5:  "Simple",
	cbor.Labeled >> 5: "Labeled",
}
