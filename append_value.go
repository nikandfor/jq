package jq

import (
	"fmt"
	"strings"

	"nikand.dev/go/cbor"
)

type (
	Arr = arr
	Obj = obj
	Raw = raw

	arr []any
	obj []any
	raw []byte
	lab struct {
		lab int
		val any
	}
)

func (b *Buffer) AppendValue(v any) (off Off) {
	return b.appendVal(v)
}

func (b *Buffer) appendVal(v any) (off Off) {
	e := b.Encoder
	var tag Tag
	var lst []any

	off = Off(len(b.B))

	switch v := v.(type) {
	case Off:
		return v
	case nil:
		return Null
	case raw:
		b.B = append(b.B, v...)
		return off
	case bool:
		if v {
			return True
		} else {
			return False
		}
	case int:
		switch v {
		case 0:
			return Zero
		case 1:
			return One
		}

		b.B = e.AppendInt(b.B, v)
		return off
	case string:
		b.B = e.AppendString(b.B, v)
		return off
	case []byte:
		b.B = e.AppendBytes(b.B, v)
		return off
	case float64:
		b.B = e.AppendFloat(b.B, v)
		return off
	case lab:
		b.B = e.AppendLabel(b.B, LabelOffset+v.lab)

		vst := Off(len(b.B))
		diff := vst - off

		val := b.appendVal(v.val)
		tag := b.Reader().Tag(val)

		if val < 0 {
			b.B = append(b.B, byte(shortToCBOR[-val]))
			return off
		}

		if !arrOrMap(tag) {
			return off
		}

		a := b.Reader().ArrayMap(val, nil)

		copy(b.B[off:], b.B[vst:])
		val -= diff
		b.B = b.B[:val]

		for i := range a {
			if a[i] >= off {
				a[i] -= diff
			}
		}

		off = Off(len(b.B))
		b.B = e.AppendLabel(b.B, v.lab)
		b.B = b.Encoder.AppendArrayMap(b.B, tag, Off(len(b.B)), a)

		return off
	case arr:
		tag = cbor.Array
		lst = []any(v)
	case obj:
		tag = cbor.Map
		lst = []any(v)
	default:
		panic(v)
	}

	var a []Off

	for _, v := range lst {
		off = b.appendVal(v)
		a = append(a, off)
	}

	off = Off(len(b.B))
	b.B = b.Encoder.AppendArrayMap(b.B, tag, off, a)

	return off
}

func (x arr) String() string {
	var b strings.Builder

	_ = b.WriteByte('[')

	for j, x := range x {
		if j != 0 {
			_, _ = b.Write([]byte{',', ' '})
		}

		_, _ = fmt.Fprintf(&b, "%v", x)
	}

	_ = b.WriteByte(']')

	return b.String()
}

func (x obj) String() string {
	var b strings.Builder

	_ = b.WriteByte('{')

	for j := 0; j < len(x); j++ {
		if j != 0 {
			_, _ = b.Write([]byte{',', ' '})
		}

		_, _ = fmt.Fprintf(&b, "%v", x[j])
		j++

		_, _ = b.Write([]byte{':', ' '})

		_, _ = fmt.Fprintf(&b, "%v", x[j])
	}

	_ = b.WriteByte('}')

	return b.String()
}
