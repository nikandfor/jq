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
	bw := b.Writer()

	if br.IsSimple(left, Null) {
		return right, more, nil
	}
	if br.IsSimple(right, Null) {
		return left, more, nil
	}

	ltag, rtag := br.Tag(left), br.Tag(right)
	lraw, rraw := br.TagRaw(left), br.TagRaw(right)

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

		f.arr = br.ArrayMap(left, f.arr[:0])
		f.arr = br.ArrayMap(right, f.arr)

		res = bw.Array(f.arr)
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

		f.arr = br.ArrayMap(left, f.arr[:0])
		l := len(f.arr)
		f.arr = br.ArrayMap(right, f.arr)
		r := l

	out:
		for i := l; i < len(f.arr); i += 2 {
			for j := 0; j < r; j += 2 {
				if b.Equal(f.arr[j], f.arr[i]) {
					f.arr[j+1] = f.arr[i+1]
					continue out
				}
			}

			f.arr[r], f.arr[r+1] = f.arr[i], f.arr[i+1]
			r += 2
		}

		res = bw.Map(f.arr[:r])
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

		return None, false, PlusError(ltag) | PlusError(rtag)<<8
	}

	return res, more, nil
}

func (f Plus) String() string {
	return fmt.Sprintf("%+v + %+v", f.L, f.R)
}

func (e PlusError) Error() string {
	return fmt.Sprintf("plus: unsupported types %v and %v", tagString(byte(e)), tagString(byte(e>>8)))
}
