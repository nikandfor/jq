package jq

import "nikand.dev/go/cbor"

type (
	Int    int
	Int64  int64
	Uint64 uint64

	Float64 float64

	Bytes       []byte
	BytesCopy   []byte
	BytesAppend []byte
)

func (f *Int) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()
	tag := br.Tag(off)
	if tag != cbor.Int && tag != cbor.Neg {
		return None, false, NewTypeError(tag, cbor.Int, cbor.Neg)
	}

	*(*int)(f) = br.Int(off)

	return None, false, nil
}

func (f *Int64) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()
	tag := br.Tag(off)
	if tag != cbor.Int && tag != cbor.Neg {
		return None, false, NewTypeError(tag, cbor.Int, cbor.Neg)
	}

	*(*int64)(f) = br.Signed(off)

	return None, false, nil
}

func (f *Uint64) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()
	tag := br.Tag(off)
	if tag != cbor.Int {
		return None, false, NewTypeError(tag, cbor.Int)
	}

	*(*uint64)(f) = br.Unsigned(off)

	return None, false, nil
}

func (f *Float64) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()
	tag := br.Tag(off)
	if tag != cbor.Int && tag != cbor.Neg {
		return None, false, NewTypeError(tag, cbor.Int, cbor.Neg)
	}

	*(*float64)(f) = br.Float(off)

	return None, false, nil
}

func (f *Bytes) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()
	tag := br.Tag(off)
	if tag != cbor.Bytes && tag != cbor.String {
		return None, false, NewTypeError(tag, cbor.Bytes, cbor.String)
	}

	*(*[]byte)(f) = br.Bytes(off)

	return None, false, nil
}

func (f *BytesCopy) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	*f = (*f)[:0]

	return (*BytesAppend)(f).ApplyTo(b, off, next)
}

func (f *BytesAppend) ApplyTo(b *Buffer, off Off, next bool) (Off, bool, error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()
	tag := br.Tag(off)
	if tag != cbor.Bytes && tag != cbor.String {
		return None, false, NewTypeError(tag, cbor.Bytes, cbor.String)
	}

	*f = append((*f), br.Bytes(off)...)

	return None, false, nil
}
