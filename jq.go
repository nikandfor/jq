package jq

import (
	"errors"
	"fmt"
	"io"

	"nikand.dev/go/cbor"
)

type (
	Filter interface {
		ApplyTo(b *Buffer, off int, next bool) (int, bool, error)
	}

	FuncFilter func(b *Buffer, off int, next bool) (int, bool, error)

	Dot     struct{}
	Empty   struct{}
	Literal []byte

	Dumper struct {
		Writer  io.Writer
		Decoder Decoder

		b   []byte
		arr []int
	}
)

const (
	_ = -iota
	None
	Null
	True
	False
	Zero
	One
)

var ErrType = errors.New("type error")

func (f FuncFilter) ApplyTo(b *Buffer, off int, next bool) (int, bool, error) {
	return f(b, off, next)
}

func (f Dot) ApplyTo(b *Buffer, off int, next bool) (int, bool, error) {
	if next {
		return None, false, nil
	}

	return off, false, nil
}

func (f Empty) ApplyTo(b *Buffer, off int, next bool) (int, bool, error) {
	return None, false, nil
}

func (f Literal) ApplyTo(b *Buffer, off int, next bool) (int, bool, error) {
	if next {
		return None, false, nil
	}

	return b.Writer().Raw(f), false, nil
}

func Dump(b []byte) string {
	return (&Dumper{}).Dump(b)
}

func DumpBuffer(b *Buffer) string {
	d := Dumper{}
	d.dumpBuffer(b)
	return string(d.b)
}

func (d *Dumper) ApplyTo(b *Buffer, off int, next bool) (int, bool, error) {
	if next {
		return None, false, nil
	}

	if d.Writer == nil {
		return off, false, nil
	}

	d.dumpBuffer(b)

	_, err := d.Writer.Write(d.b)
	if err != nil {
		return off, false, err
	}

	return off, false, nil
}

func (d *Dumper) Dump(b []byte) string {
	d.b = d.b[:0]

	d.dump(b, -1, 0)
	d.b = fmt.Appendf(d.b, "%06x\n", len(b))

	return string(d.b)
}

func (d *Dumper) DumpBuffer(b *Buffer) string {
	d.dumpBuffer(b)
	return string(d.b)
}

func (d *Dumper) dumpBuffer(b *Buffer) {
	d.b = append(d.b, "rbuf\n"...)
	d.dump(b.R, 0, 0)
	d.b = append(d.b, "wbuf\n"...)
	d.dump(b.W, len(b.R), 0)

	d.b = fmt.Appendf(d.b, "%06x\n", len(b.R)+len(b.W))
}

func (d *Dumper) dump(b []byte, base, depth int) {
	const spaces = "                    "

	for i := 0; i >= 0 && i < len(b); {
		st := i
		tag := d.Decoder.TagOnly(b, i)

		//	log.Printf("dump loop %x", i)

		if base >= 0 {
			d.b = fmt.Appendf(d.b, "%06x  ", base+i)
		}

		d.b = fmt.Appendf(d.b, "%06x  %s", i, spaces[:depth])

		switch tag {
		case cbor.Int:
			var v uint64
			v, i = d.Decoder.Unsigned(b, i)

			d.b = fmt.Appendf(d.b, "% 02x  %d\n", b[st:i], v)
		case cbor.Neg:
			var v uint64
			v, i = d.Decoder.Unsigned(b, i)

			d.b = fmt.Appendf(d.b, "% 02x  -%d\n", b[st:i], v)
		case cbor.Bytes:
			_, i = d.Decoder.Bytes(b, i)

			d.b = fmt.Appendf(d.b, "% 02x\n", b[st:i])
		case cbor.String:
			var v []byte
			v, i = d.Decoder.Bytes(b, i)

			d.b = fmt.Appendf(d.b, "% 02x  ", b[st:i])

			var qq, bq bool

			for _, c := range v {
				qq = qq || c == '"'
				bq = bq || c == '`'
			}

			if qq && !bq {
				d.b = append(d.b, '`')
				d.b = append(d.b, v...)
				d.b = append(d.b, '`', '\n')
			} else {
				d.b = fmt.Appendf(d.b, "%q\n", v)
			}
		case cbor.Simple:
			i = d.Decoder.Skip(b, i)
			if i < 0 || i > len(b) {
				d.b = fmt.Appendf(d.b, "broken string %x:%x / %x", st, i, len(b))
				break
			}

			d.b = fmt.Appendf(d.b, "% 02x\n", b[st:i])
		case cbor.Labeled:
			_, _, i = d.Decoder.CBOR.Tag(b, i)

			d.b = fmt.Appendf(d.b, "% 02x\n", b[st:i])
			depth += 4
			continue
		case cbor.Array, cbor.Map:
			bb := 0
			if base >= 0 {
				bb = base
			}

			_, l, s, end := d.Decoder.TagArrayMap(b, i)
			d.arr, i = d.Decoder.ArrayMap(b, bb, i, d.arr[:0])

			d.b = fmt.Appendf(d.b, "% 02x  l %2d  s %d\n", b[st:end], l, s)

			for j, off := range d.arr {
				if base >= 0 {
					d.b = fmt.Appendf(d.b, "%s", spaces[:8])
				}
				d.b = fmt.Appendf(d.b, "%s", spaces[:8+depth+4])
				d.b = fmt.Appendf(d.b, "%3x: %6x\n", j, off)
			}
		default:
			panic(tag)
		}

		if depth > 0 {
			depth -= 4
		}
	}
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
