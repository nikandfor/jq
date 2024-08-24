package jq

type (
	Buffer struct {
		r, w []byte

		Decoder Decoder
		Encoder Encoder
	}

	BufferReader struct {
		*Buffer
	}
	BufferWriter struct {
		*Buffer
	}
)

func NewBuffer(r []byte) *Buffer {
	return &Buffer{r: r}
}

func MakeBuffer(r []byte) Buffer {
	return Buffer{r: r}
}

func (b *Buffer) Reset(r []byte) {
	b.r = r
	b.w = b.w[:0]
}

func (b *Buffer) Reader() BufferReader { return BufferReader{b} }
func (b *Buffer) Writer() BufferWriter { return BufferWriter{b} }

func (b BufferReader) Tag(off int) byte {
	tag, _ := b.Decoder.Tag(b.buf(off))
	return tag
}

func (b BufferReader) Raw(off int) []byte {
	raw, _ := b.Decoder.Raw(b.buf(off))
	return raw
}

func (b BufferReader) Bytes(off int) []byte {
	s, _ := b.Decoder.Bytes(b.buf(off))
	return s
}

func (b BufferReader) ArrayMapIndex(off, index int) (k, v int) {
	buf, base, eoff := b.bbuf(off)
	return b.Decoder.ArrayMapIndex(buf, base, eoff, index)
}

func (b BufferReader) ArrayMap(off int, arr []int) []int {
	buf, base, eoff := b.bbuf(off)
	arr, _ = b.Decoder.ArrayMap(buf, base, eoff, arr)
	return arr
}

func (b BufferWriter) Offset() int {
	return len(b.r) + len(b.w)
}

func (b BufferWriter) Reset(off int) {
	b.w = b.w[:off-len(b.r)]
}

func (b BufferWriter) ResetIfErr(off int, errp *error) {
	if *errp == nil {
		return
	}

	b.Reset(off)
}

func (b BufferWriter) Raw(raw []byte) int {
	off := b.Offset()
	b.w = append(b.w, raw...)

	return off
}

func (b BufferWriter) Array(arr []int) int {
	off := b.Offset()
	b.w = b.Encoder.AppendArray(b.w, off, arr)
	return off
}

func (b BufferWriter) Map(arr []int) int {
	off := b.Offset()
	b.w = b.Encoder.AppendMap(b.w, off, arr)
	return off
}

func (b *Buffer) buf(off int) ([]byte, int) {
	if off < len(b.r) {
		return b.r, off
	}

	return b.w, off - len(b.r)
}

func (b *Buffer) bbuf(off int) ([]byte, int, int) {
	if off <= len(b.r) {
		return b.r, 0, off
	}

	return b.w, len(b.r), off - len(b.r)
}
