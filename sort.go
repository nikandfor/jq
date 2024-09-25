package jq

import (
	"bytes"
	"fmt"
	"sort"

	"nikand.dev/go/cbor"
)

type (
	arrSort struct {
		b *Buffer
		v []Off
	}

	mapSort struct {
		b *Buffer
		v []Off
	}
)

func (b *Buffer) SortArray(v []Off) {
	sort.Sort(arrSort{b, v})
}

func (b *Buffer) SortMap(v []Off) {
	sort.Sort(mapSort{b, v})
}

func (x arrSort) Len() int      { return len(x.v) }
func (x mapSort) Len() int      { return len(x.v) / 2 }
func (x arrSort) Swap(i, j int) { x.v[i], x.v[j] = x.v[j], x.v[i] }
func (x mapSort) Swap(i, j int) {
	i *= 2
	j *= 2

	x.v[i], x.v[i+1], x.v[j], x.v[j+1] = x.v[j], x.v[j+1], x.v[i], x.v[i+1]
}

func (x arrSort) Less(i, j int) bool {
	return x.b.Compare(x.v[i], x.v[j]) < 0
}

func (x mapSort) Less(i, j int) bool {
	i *= 2
	j *= 2

	return x.b.Compare(x.v[i], x.v[j]) < 0
}

const (
	less    = -1
	greater = 1
)

func (b *Buffer) Compare(x, y Off) int {
	if b.Equal(x, y) {
		return 0
	}

	if r := b.cmpShort(x, y, Null); r != 0 {
		return r
	}
	if r := b.cmpShort(x, y, False); r != 0 {
		return r
	}
	if r := b.cmpShort(x, y, True); r != 0 {
		return r
	}

	br := b.Reader()

	if r := b.cmpTags(x, y); r != 0 {
		return r
	}

	if cbor.IsNum(br.TagRaw(x)) {
		return b.cmpNum(x, y)
	}

	xt, yt := br.Tag(x), br.Tag(y)

	if xt == cbor.Bytes || xt == cbor.String {
		return b.cmpBytes(x, y)
	}

	if xt == cbor.Array || yt == cbor.Map {
		return b.cmpArrayMap(x, y)
	}

	panic(fmt.Sprintf("%+v %+v", x, y))
}

func (b *Buffer) cmpNum(x, y Off) int {
	br := b.Reader()

	if cbor.IsInt(br.Tag(x)) {
		tag := br.Tag(x)
		x := br.Unsigned(x)
		y := br.Unsigned(y)

		if x == y {
			return 0
		}

		if (x < y) == (tag == cbor.Int) {
			return less
		} else {
			return greater
		}
	}

	xf := br.Float(x)
	yf := br.Float(y)

	if xf == yf {
		return 0
	}

	if xf < yf {
		return less
	} else {
		return greater
	}
}

func (b *Buffer) cmpBytes(x, y Off) int {
	br := b.Reader()

	xv := br.Bytes(x)
	yv := br.Bytes(y)

	return bytes.Compare(xv, yv)
}

func (b *Buffer) cmpArrayMap(x, y Off) int {
	br := b.Reader()

	val := csel(br.Tag(x) == cbor.Map, 2, 1)

	b.arr = br.ArrayMap(x, b.arr[:0])
	l := len(b.arr)
	b.arr = br.ArrayMap(y, b.arr)

	for i := 0; i < l && i < len(b.arr[l:]); i++ {
		if r := b.Compare(b.arr[i], b.arr[l+i]); r != 0 {
			return r
		}

		i += val
	}

	if l == len(b.arr[l:]) {
		return 0
	}

	if l < len(b.arr[l:]) {
		return less
	} else {
		return greater
	}
}

func (b *Buffer) cmpShort(x, y, v Off) int {
	br := b.Reader()

	if br.IsSimple(x, v) {
		return less
	}
	if br.IsSimple(y, v) {
		return greater
	}

	return 0
}

func (b *Buffer) cmpTags(xoff, yoff Off) int {
	br := b.Reader()
	x, y := br.Tag(xoff), br.Tag(yoff)
	xr, yr := br.TagRaw(xoff), br.TagRaw(yoff)

	if x == y {
		return 0
	}

	if cbor.IsNum(xr) && cbor.IsNum(yr) {
		if cbor.IsFloat(xr) || cbor.IsFloat(yr) {
			return 0
		}

		if x == cbor.Neg {
			return less
		} else {
			return greater
		}
	}

	if cbor.IsNum(xr) {
		return less
	}
	if cbor.IsNum(yr) {
		return greater
	}

	if x == cbor.Bytes {
		return less
	}
	if y == cbor.Bytes {
		return greater
	}

	if x == cbor.String {
		return less
	}
	if y == cbor.String {
		return greater
	}

	if x == cbor.Array {
		return less
	}
	if y == cbor.Array {
		return greater
	}

	if x == cbor.Map {
		return less
	}
	if y == cbor.Map {
		return greater
	}

	panic(x)
}
