package jq

import (
	"fmt"
	"strconv"

	"nikand.dev/go/cbor"
)

type (
	Convert struct {
		Type ConvertType
	}

	ConvertType Tag
)

const (
	_ ConvertType = iota
	ToBytes
	ToString
	ToInt
	ToFloat
)

func NewConvert(to ConvertType) Convert {
	return Convert{Type: to}
}

func (f Convert) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()
	bw := b.Writer()
	tag := br.Tag(off)
	raw := br.TagRaw(off)

	var typ cbor.Tag

	switch f.Type {
	case ToInt:
		typ = cbor.Int
	case ToBytes:
		typ = cbor.Bytes
	case ToString:
		typ = cbor.String
	}

	switch f.Type {
	case ToBytes, ToString:
		switch {
		case tag == typ:
			res = off
		case tag == cbor.String || tag == cbor.Bytes:
			s := br.Bytes(off)
			res = bw.TagBytes(typ, s)
		case cbor.IsInt(raw) || cbor.IsFloat(raw):
			var x any
			res = bw.Off()
			ww := bw.StringWriter(typ)

			neg := csel(tag == cbor.Neg, "-", "")

			switch tag {
			case cbor.Int, cbor.Neg:
				x = br.Unsigned(off)
			default:
				x = br.Float(off)
			}

			fmt.Fprintf(ww, "%s%v", neg, x)
		default:
			return None, false, NewTypeError(raw)
		}
	case ToInt:
		switch {
		case cbor.IsInt(raw):
			res = off
		case cbor.IsFloat(raw):
			x := int64(br.Float(off))
			res = bw.Int64(x)
		case tag == cbor.Bytes, tag == cbor.String:
			s := br.Bytes(off)
			x, err := strconv.ParseInt(string(s), 10, 64)
			if err != nil {
				return None, false, err
			}

			res = bw.Int64(x)
		default:
			return None, false, NewTypeError(raw)
		}
	case ToFloat:
		switch {
		case cbor.IsFloat(raw):
			res = off
		case tag == cbor.Int:
			x := br.Unsigned(off)
			res = bw.Float(float64(x))
		case tag == cbor.Neg:
			x := br.Signed(off)
			res = bw.Float(float64(x))
		case tag == cbor.Bytes, tag == cbor.String:
			s := br.Bytes(off)
			x, err := strconv.ParseFloat(string(s), 64)
			if err != nil {
				return None, false, err
			}

			res = bw.Float(x)
		default:
			return None, false, NewTypeError(raw)
		}
	default:
		return None, false, NewTypeError(typ)
	}

	return res, false, nil
}

func (f Convert) String() string {
	switch f.Type {
	case ToBytes:
		return "toBytes"
	case ToString:
		return "toString"
	case ToInt:
		return "toInt"
	case ToFloat:
		return "toFloat"
	default:
		return fmt.Sprintf("%x", int(f.Type))
	}
}
