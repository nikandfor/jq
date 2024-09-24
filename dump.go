package jq

import (
	"fmt"
	"io"

	"nikand.dev/go/cbor"
)

type (
	Dumper struct {
		Writer  io.Writer
		Decoder Decoder

		Base int

		b   []byte
		arr []Off
	}
)

func DumpBytes(b []byte) string {
	return (&Dumper{Base: -1}).DumpBytes(b)
}

func Dump(b *Buffer) string {
	d := Dumper{Base: -1}
	d.dumpBuffer(b)
	return string(d.b)
}

func NewDumper(w io.Writer) *Dumper { return &Dumper{Writer: w, Base: -1} }

func (d *Dumper) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
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

func (d *Dumper) DumpBytes(b []byte) string {
	d.b = d.b[:0]

	d.dump(b, 0)

	if d.Base >= 0 {
		d.Base += len(b) // TODO: base
		d.b = fmt.Appendf(d.b, "%06x  ", d.Base)
	}

	d.b = fmt.Appendf(d.b, "%06x\n", len(b))

	return string(d.b)
}

func (d *Dumper) Dump(b *Buffer) string {
	d.dumpBuffer(b)
	return string(d.b)
}

func (d *Dumper) dumpBuffer(b *Buffer) {
	d.dump(b.B, 0)

	if d.Base >= 0 {
		d.Base += len(b.B)
		d.b = fmt.Appendf(d.b, "%06x  ", d.Base)
	}

	d.b = fmt.Appendf(d.b, "%06x\n", len(b.B))
}

func (d *Dumper) dump(b []byte, depth int) {
	const spaces = "                    "

	defer func() {
		p := recover()
		if p == nil {
			return
		}

		defer panic(p)

		d.b = fmt.Appendf(d.b, "panic: %v\n", p)
	}()

	for i := 0; i >= 0 && i < len(b); {
		st := i
		tag := d.Decoder.TagOnly(b, i)

		//	log.Printf("dump loop %x", i)

		if d.Base >= 0 {
			d.b = fmt.Appendf(d.b, "%06x  ", d.Base+i)
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
			d.b = d.appendQStr(d.b, v)
		case cbor.Simple:
			i = d.Decoder.Skip(b, i)

			d.b = fmt.Appendf(d.b, "% 02x\n", b[st:i])
		case cbor.Labeled:
			_, _, i = d.Decoder.CBOR.Tag(b, i)

			d.b = fmt.Appendf(d.b, "% 02x\n", b[st:i])
			depth += 4
			continue
		case cbor.Array, cbor.Map:
			_, l, s, end := d.Decoder.TagArrayMap(b, i)
			d.arr, i = d.Decoder.ArrayMap(b, i, d.arr[:0])

			d.b = fmt.Appendf(d.b, "% 02x  l %2d  s %d\n", b[st:end], l, s)

			for j, off := range d.arr {
				if d.Base >= 0 {
					d.b = fmt.Appendf(d.b, "%s", spaces[:8])
				}
				d.b = fmt.Appendf(d.b, "%s", spaces[:8+depth+4])
				d.b = fmt.Appendf(d.b, "%3x: %6x\n", j, int64(off))
			}
		default:
			panic(tag)
		}

		if depth > 0 {
			depth -= 4
		}
	}
}

func (d *Dumper) encodeString(w []byte, b *Buffer, off Off) ([]byte, error) {
	tag := b.Reader().Tag(off)
	r, st := b.Buf(off)

	w = append(w, '"')
	w, i := d.encStr(w, r, st)
	w = append(w, '"')
	if i < 0 {
		return w, NewTypeError(tag, cbor.Bytes, cbor.String)
	}

	return w, nil
}

func (d *Dumper) encStr(w, r []byte, i int) ([]byte, int) {
	tag, sub, i := d.Decoder.CBOR.Tag(r, i)
	l := int(sub)
	if tag != cbor.Bytes && tag != cbor.String {
		return w, -1
	}
	if l >= 0 {
		return d.appendQStr(w, r[i:i+l]), i + l
	}

	for !d.Decoder.CBOR.Break(r, &i) {
		w, i = d.encStr(w, r, i)
		if i < 0 {
			return w, i
		}
	}

	return w, i
}

func (d *Dumper) appendQStr(w, v []byte) []byte {
	var qq, bq bool

	for _, c := range v {
		qq = qq || c == '"'
		bq = bq || c == '`'
	}

	if qq && !bq {
		w = append(w, '`')
		w = append(w, v...)
		w = append(w, '`', '\n')
	} else {
		w = fmt.Appendf(w, "%q\n", v)
	}

	return w
}
