package jq

import "nikand.dev/go/cbor"

func (b BufferWriter) Off() Off {
	return Off(len(b.R) + len(b.W))
}

func (b BufferWriter) Reset(off Off) {
	b.W = b.W[:int(off)-len(b.R)]
}

func (b BufferWriter) ResetIfErr(off Off, errp *error) {
	if *errp == nil {
		return
	}

	b.Reset(off)
}

func (b BufferWriter) Raw(raw []byte) Off {
	off := b.Off()
	b.W = append(b.W, raw...)

	return off
}

func (b BufferWriter) Array(arr []Off) Off {
	if len(arr) == 0 {
		return EmptyArray
	}

	off := b.Off()
	b.W = b.Encoder.AppendArray(b.W, off, arr)
	return off
}

func (b BufferWriter) Map(arr []Off) Off {
	off := b.Off()
	b.W = b.Encoder.AppendMap(b.W, off, arr)
	return off
}

func (b BufferWriter) ArrayMap(tag byte, arr []Off) Off {
	if tag == cbor.Array && len(arr) == 0 {
		return EmptyArray
	}

	off := b.Off()
	b.W = b.Encoder.AppendArrayMap(b.W, tag, off, arr)
	return off
}

func (b BufferWriter) String(v string) Off {
	if v == "" {
		return EmptyString
	}

	off := b.Off()
	b.W = b.Encoder.AppendString(b.W, v)
	return off
}

func (b BufferWriter) Bytes(v []byte) Off {
	off := b.Off()
	b.W = b.Encoder.AppendBytes(b.W, v)
	return off
}

func (b BufferWriter) TagString(tag byte, v string) Off {
	if tag == cbor.String && v == "" {
		return EmptyString
	}

	off := b.Off()
	b.W = b.Encoder.AppendTagString(b.W, tag, v)
	return off
}

func (b BufferWriter) TagBytes(tag byte, v []byte) Off {
	if tag == cbor.String && len(v) == 0 {
		return EmptyString
	}

	off := b.Off()
	b.W = b.Encoder.AppendTagBytes(b.W, tag, v)
	return off
}

func (b BufferWriter) Int(v int) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.W = b.Encoder.AppendInt(b.W, v)
	return off
}

func (b BufferWriter) Int64(v int64) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.W = b.Encoder.AppendInt64(b.W, v)
	return off
}

func (b BufferWriter) Uint(v uint) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.W = b.Encoder.AppendUint(b.W, v)
	return off
}

func (b BufferWriter) Uint64(v uint64) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.W = b.Encoder.AppendUint64(b.W, v)
	return off
}

func (b BufferWriter) Float(v float64) Off {
	off := b.Off()
	b.W = b.Encoder.AppendFloat(b.W, v)
	return off
}

func (b BufferWriter) Float32(v float32) Off {
	off := b.Off()
	b.W = b.Encoder.AppendFloat32(b.W, v)
	return off
}
