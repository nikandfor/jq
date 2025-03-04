package jq

import (
	"io"

	"nikand.dev/go/cbor"
	"nikand.dev/go/skip"
)

func (b BufferWriter) Off() Off {
	return Off(len(b.B))
}

func (b BufferWriter) Reset(off Off) {
	if off < 0 {
		return
	}

	b.B = b.B[:off]
}

func (b BufferWriter) ResetIfErr(off Off, errp *error) {
	if *errp == nil {
		return
	}

	b.Reset(off)
}

func (b BufferWriter) Raw(raw []byte) Off {
	if len(raw) == 1 {
		off := cborToShort[raw[0]]
		if off < 0 {
			return off
		}
	}

	if off := b.GetStatic(raw); off != None {
		return off
	}

	off := b.Off()
	b.B = append(b.B, raw...)

	b.AddStatic(off)

	return off
}

func (b BufferWriter) Array(arr []Off) Off {
	if len(arr) == 0 {
		return EmptyArray
	}

	off := b.Off()
	b.B = b.Encoder.AppendArray(b.B, off, arr)
	return off
}

func (b BufferWriter) Map(arr []Off) Off {
	off := b.Off()
	b.B = b.Encoder.AppendMap(b.B, off, arr)
	return off
}

func (b BufferWriter) ArrayMap(tag Tag, arr []Off) Off {
	if tag == cbor.Array && len(arr) == 0 {
		return EmptyArray
	}

	off := b.Off()
	b.B = b.Encoder.AppendArrayMap(b.B, tag, off, arr)
	return off
}

func (b BufferWriter) String(v string) Off {
	if v == "" {
		return EmptyString
	}

	off := b.Off()
	b.B = b.Encoder.AppendString(b.B, v)

	if static := b.GetStatic(b.B[off:]); static != None {
		b.Reset(off)
		return static
	}

	b.AddStatic(off)

	return off
}

func (b BufferWriter) Bytes(v []byte) Off {
	off := b.Off()
	b.B = b.Encoder.AppendBytes(b.B, v)

	if static := b.GetStatic(b.B[off:]); static != None {
		b.Reset(off)
		return static
	}

	b.AddStatic(off)

	return off
}

func (b BufferWriter) TagString(tag Tag, v string) Off {
	if tag == cbor.String && v == "" {
		return EmptyString
	}

	off := b.Off()
	b.B = b.Encoder.AppendTagString(b.B, tag, v)

	if static := b.GetStatic(b.B[off:]); static != None {
		b.Reset(off)
		return static
	}

	b.AddStatic(off)

	return off
}

func (b BufferWriter) TagBytes(tag Tag, v []byte) Off {
	if tag == cbor.String && len(v) == 0 {
		return EmptyString
	}

	off := b.Off()
	b.B = b.Encoder.AppendTagBytes(b.B, tag, v)

	if static := b.GetStatic(b.B[off:]); static != None {
		b.Reset(off)
		return static
	}

	b.AddStatic(off)

	return off
}

func (b BufferWriter) Int(v int) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.B = b.Encoder.AppendInt(b.B, v)
	return off
}

func (b BufferWriter) Int64(v int64) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.B = b.Encoder.AppendInt64(b.B, v)
	return off
}

func (b BufferWriter) Uint(v uint) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.B = b.Encoder.AppendUint(b.B, v)
	return off
}

func (b BufferWriter) Uint64(v uint64) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.B = b.Encoder.AppendUint64(b.B, v)
	return off
}

func (b BufferWriter) TagUnsigned(tag Tag, v uint64) Off {
	if v == 0 || v == 1 {
		return Zero - Off(v)
	}

	off := b.Off()
	b.B = b.Encoder.AppendTagUnsigned(b.B, tag, v)
	return off
}

func (b BufferWriter) Float(v float64) Off {
	off := b.Off()
	b.B = b.Encoder.AppendFloat(b.B, v)
	return off
}

func (b BufferWriter) Float32(v float32) Off {
	off := b.Off()
	b.B = b.Encoder.AppendFloat32(b.B, v)
	return off
}

func (b BufferWriter) Label(lab int) Off {
	off := b.Off()
	b.B = b.Encoder.AppendLabel(b.B, lab)
	return off
}

type bufferIOWriter struct {
	b   BufferWriter
	tag Tag
}

func (b BufferWriter) StringWriter(tag Tag) io.Writer {
	return bufferIOWriter{b, tag}
}

func (b BufferWriter) RawWriter() io.Writer {
	return bufferIOWriter{b, 0}
}

func (w bufferIOWriter) Write(p []byte) (int, error) {
	if w.tag == 0 {
		_ = w.b.Raw(p)
	} else {
		_ = w.b.TagBytes(w.tag, p)
	}

	return len(p), nil
}

func (b BufferWriter) GetStatic(raw []byte) Off {
	if !b.Flags.Is(BufferStatic) {
		return None
	}

	if len(raw) <= 2 {
		return None
	}

	for _, off := range b.static {
		if len(raw) <= 4 && b.Off()-off >= 0x10000-offReserve {
			continue
		}

		if !skip.Prefix(b.B[off:], raw) {
			continue
		}

		return off
	}

	return None
}

func (b BufferWriter) AddStatic(off Off) {
	if !b.Flags.Is(BufferStatic) {
		return
	}

	b.static = append(b.static, off)
}
