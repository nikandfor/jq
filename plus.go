package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Plus struct {
		L, R Filter

		binop Binop

		arr []Off
	}

	PlusError int
)

func NewPlus(l, r Filter) *Plus {
	return &Plus{L: l, R: r}
}

func (f *Plus) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	left, right, more, err := f.binop.ApplyTo(b, off, next, f.L, f.R)
	if err != nil || left == None || right == None {
		return None, more, err
	}

	br := b.Reader()

	if br.IsSimple(left, Null) {
		return right, more, nil
	}
	if br.IsSimple(right, Null) {
		return left, more, nil
	}

	res, err = plus(b, left, right)
	if err != nil {
		return None, false, err
	}

	return res, more, nil
}

func plus(b *Buffer, left, right Off) (res Off, err error) {
	br := b.Reader()
	bw := b.Writer()

	if left == Null {
		return right, nil
	}
	if right == Null {
		return left, nil
	}

	ltag, rtag := br.Tag(left), br.Tag(right)
	lraw, rraw := br.TagRaw(left), br.TagRaw(right)

	arrbase := len(b.arr)
	defer func() { b.arr = b.arr[:arrbase] }()

	switch {
	case ltag == cbor.Int && rtag == cbor.Int:
		l := br.Unsigned(left)
		r := br.Unsigned(right)

		res = bw.Uint64(l + r)
	case (ltag == cbor.Int || ltag == cbor.Neg) && (rtag == cbor.Int || rtag == cbor.Neg):
		l := br.Signed(left)
		r := br.Signed(right)

		res = bw.Int64(l + r)
	case (ltag == cbor.Bytes || ltag == cbor.String) && (rtag == cbor.Bytes || rtag == cbor.String):
		l := br.Bytes(left)
		r := br.Bytes(right)

		if len(l)+len(r) == 0 && ltag == cbor.String {
			res = EmptyString
			break
		}
		if len(r) == 0 {
			res = left
			break
		}
		if len(l) == 0 && ltag == rtag {
			res = right
			break
		}

		res = bw.Off()

		b.B = b.Encoder.CBOR.AppendTag(b.B, ltag, len(l)+len(r))
		b.B = append(b.B, l...)
		b.B = append(b.B, r...)
	case ltag == cbor.Array && rtag == cbor.Array:
		ll := br.ArrayMapLen(left)
		rl := br.ArrayMapLen(right)

		if ll == 0 && rl == 0 {
			res = EmptyArray
			break
		}
		if ll == 0 {
			res = right
			break
		}
		if rl == 0 {
			res = left
			break
		}

		b.arr = br.ArrayMap(left, b.arr)
		b.arr = br.ArrayMap(right, b.arr)

		res = bw.Array(b.arr[arrbase:])
	case ltag == cbor.Map && rtag == cbor.Map:
		ll := br.ArrayMapLen(left)
		rl := br.ArrayMapLen(right)

		if ll == 0 {
			res = right
			break
		}
		if rl == 0 {
			res = left
			break
		}

		b.arr = br.ArrayMap(left, b.arr)
		l := len(b.arr)
		b.arr = br.ArrayMap(right, b.arr)
		r := l

	out:
		for i := l; i < len(b.arr); i += 2 {
			for j := 0; j < r; j += 2 {
				if b.Equal(b.arr[j], b.arr[i]) {
					b.arr[j+1] = b.arr[i+1]
					continue out
				}
			}

			b.arr[r], b.arr[r+1] = b.arr[i], b.arr[i+1]
			r += 2
		}

		res = bw.Map(b.arr[arrbase:r])
	default:
		fmin := cbor.Simple | cbor.Float8
		fmax := cbor.Simple | cbor.Float64

		lfloat := lraw >= fmin && lraw <= fmax
		rfloat := rraw >= fmin && rraw <= fmax
		lint := ltag == cbor.Int || ltag == cbor.Neg
		rint := rtag == cbor.Int || rtag == cbor.Neg

		if (lfloat || lint) && (rfloat || rint) {
			var l, r float64

			if lfloat {
				l = br.Float(left)
			} else {
				l = float64(br.Signed(left))
			}

			if rfloat {
				r = br.Float(right)
			} else {
				r = float64(br.Signed(right))
			}

			res = bw.Float(l + r)

			break
		}

		return None, PlusError(ltag) | PlusError(rtag)<<8
	}

	return res, nil
}

func (f Plus) String() string {
	return fmt.Sprintf("%+v + %+v", f.L, f.R)
}

func (e PlusError) Error() string {
	return fmt.Sprintf("plus: unsupported types %v and %v", tagString(Tag(e)), tagString(Tag(e>>8)))
}
