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
	b.B, off = appendValBuf(b.B, 0, v)
	if off < 0 {
		return off
	}

	return off
}

func appendValBuf(w []byte, base int, v any) ([]byte, Off) {
	var e Encoder
	var a []Off
	var tag Tag
	var lst []any

	off := Off(base + len(w))

	switch v := v.(type) {
	case Off:
		return w, v
	case nil:
		return w, Null
	case raw:
		return append(w, v...), off
	case bool:
		if v {
			return w, True
		} else {
			return w, False
		}
	case int:
		switch v {
		case 0:
			return w, Zero
		case 1:
			return w, One
		}

		w = e.CBOR.AppendInt(w, v)
		return w, off
	case string:
		w = e.CBOR.AppendString(w, v)
		return w, off
	case []byte:
		w = e.CBOR.AppendBytes(w, v)
		return w, off
	case float64:
		w = e.CBOR.AppendFloat(w, v)
		return w, off
	case lab:
		w = e.CBOR.AppendLabeled(w, v.lab)
		w, _ = appendValBuf(w, base, v.val)
		return w, off
	case arr:
		tag = cbor.Array
		lst = []any(v)
	case obj:
		tag = cbor.Map
		lst = []any(v)
	default:
		panic(v)
	}

	for _, v := range lst {
		w, off = appendValBuf(w, base, v)
		a = append(a, off)
	}

	off = Off(base + len(w))
	w = e.AppendArrayMap(w, tag, off, a)

	return w, off
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
