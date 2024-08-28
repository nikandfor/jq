package jq

import "nikand.dev/go/cbor"

type (
	Arr  = arr
	Code = code
	Obj  = obj
	Raw  = raw

	code int
	arr  []any
	obj  []any
	raw  []byte
	lab  struct {
		lab int
		val any
	}
)

func (b *Buffer) AppendValue(v any) (off int) {
	return b.appendVal(v)
}

func (b *Buffer) appendVal(v any) (off int) {
	b.W, off = appendValBuf(b.W, len(b.R), v)
	if off < 0 {
		return off
	}

	return off
}

func appendValBuf(w []byte, base int, v any) ([]byte, int) {
	var e Encoder
	var a []int
	var tag byte
	var lst []any

	off := base + len(w)

	switch v := v.(type) {
	case code:
		return w, int(v)
	case Off:
		return w, int(v)
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

	off = base + len(w)
	w = e.AppendArrayMap(w, tag, off, a)

	return w, off
}
