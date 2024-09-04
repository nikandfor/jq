package jq

import (
	"errors"
	"fmt"
	"strings"

	"nikand.dev/go/cbor"
)

type (
	Query struct {
		Path []any

		IgnoreTypeError bool

		stack []indexState
		arr   []int
	}

	indexState struct {
		off        int
		st, end, i int
		val        int
	}
)

var ErrUnsupportedIndexKey = errors.New("unsupported index key")

func NewQuery(p ...any) *Query {
	return &Query{Path: query(p)}
}

func query(p []any) []any {
	for i := range p {
		switch p[i].(type) {
		case string, int, Iter:
		case *Iter:
			p[i] = Iter{}
		default:
			panic(p[i])
		}
	}

	return p
}

func (f *Query) ApplyToGetPath(b *Buffer, off int, next bool, base Path) (res int, path Path, more bool, err error) {
	res, more, err = f.ApplyTo(b, off, next)
	if err != nil {
		return off, base, false, err
	}

	for _, st := range f.stack {
		index := st.i - st.st
		if st.val == 1 {
			index /= 2
		}

		base = append(base, PathSeg{
			Off:   st.off,
			Index: index,
		})
	}

	return res, base, more, nil
}

func (f *Query) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	br := b.Reader()

	if len(f.Path) == 0 {
		return None, false, nil
	}

	//	defer func(off int, next bool) {
	//		log.Printf("index %x %v -> %x %v %v   %v", off, next, res, more, err, f.stack)
	//	}(off, next)

	if !next {
		f.init(off)
	}

	fi := len(f.Path)

back:
	for {
		fi = f.back(fi, &next)
		//	log.Printf("index %x back  %v", fi, f.stack)
		if fi < 0 {
			return None, false, nil // TODO
		}

		off = f.stack[fi].off

		for ; fi < len(f.Path); fi++ {
			f.stack[fi].off = off
			//	log.Printf("index %x step  %v", fi, f.stack)

			tag := br.Tag(off)
			k := f.Path[fi]

			switch k := k.(type) {
			case int:
				if off == Null {
					continue
				}

				if tag != cbor.Map && tag != cbor.Array {
					return off, false, ErrType
				}

				_, off = br.ArrayMapIndex(off, k)
				f.stack[fi].i = k
			case string:
				if off == Null {
					continue
				}

				if tag != cbor.Map && f.IgnoreTypeError {
					off = Null
					break
				}
				if tag != cbor.Map {
					return off, false, ErrType
				}

				off, f.stack[fi].i = f.mapKey(b, off, k)
				f.stack[fi].val = 1
			case Iter, *Iter:
				if off == Null || tag != cbor.Map && tag != cbor.Array {
					return off, false, ErrType
				}

				if f.stack[fi].end < 0 {
					val := csel(tag == cbor.Map, 1, 0)

					st := len(f.arr)
					f.arr = br.ArrayMap(off, f.arr)

					f.stack[fi] = indexState{off: off, st: st, i: st, end: len(f.arr), val: val}
				} else {
					f.stack[fi].i += 1 + f.stack[fi].val
				}

				st := f.stack[fi]

				if st.i == st.end {
					fi++
					continue back
				}

				off = f.arr[st.i+st.val]
			default:
				return off, false, ErrUnsupportedIndexKey
			}
		}

		break
	}

	more = f.back(len(f.Path), nil) >= 0

	return off, more, nil
}

func (f *Query) back(fi int, next *bool) int {
	if next != nil && !*next {
		*next = true
		return 0
	}

	for fi--; fi >= 0; fi-- {
		st := f.stack[fi]
		if st.end < 0 {
			continue
		}
		if st.i+1+st.val < st.end {
			break
		}

		f.stack[fi].end = -1
	}

	return fi
}

func (f *Query) init(root int) bool {
	f.arr = f.arr[:0]

	for cap(f.stack) < len(f.Path) {
		f.stack = append(f.stack[:cap(f.stack)], indexState{})
	}

	f.stack = f.stack[:len(f.Path)]

	for j := range f.stack {
		f.stack[j] = indexState{off: None, end: -1}
	}

	f.stack[0].off = root

	return true
}

func (f *Query) mapKey(b *Buffer, off int, key string) (res, i int) {
	br := b.Reader()
	reset := len(f.arr)
	res = Null

	f.arr = br.ArrayMap(off, f.arr)
	//	log.Printf("mapkey %x %q from %x", off, key, f.arr)

	for j := reset; j < len(f.arr); j += 2 {
		k := br.Bytes(f.arr[j])
		//	log.Printf("mapkey %3x %q -> %d %x  %q", off, key, j-reset, f.arr[j], k)
		if string(k) != key {
			continue
		}

		res = f.arr[j+1]
		i = j - reset
		break
	}

	f.arr = f.arr[:reset]

	return res, i
}

func (f *Query) String() string {
	var b strings.Builder

	//	b.WriteString("Index{")

	if len(f.Path) == 0 {
		b.WriteByte('.')
	} else if _, ok := f.Path[0].(string); !ok {
		b.WriteByte('.')
	}

	for _, p := range f.Path {
		switch p := p.(type) {
		case string:
			fmt.Fprintf(&b, ".%s", p)
		case int:
			fmt.Fprintf(&b, "[%d]", p)
		case Iter:
			fmt.Fprintf(&b, "[]")
		default:
			fmt.Fprintf(&b, "%v", p)
		}
	}

	//	b.WriteString("}")

	return b.String()
}

func (s indexState) String() string {
	return fmt.Sprintf("{off %x, st %x, i %x, end %x, val %x}", s.off, s.st, s.i, s.end, s.val)
}

func csel[T any](c bool, t, f T) T {
	if c {
		return t
	}

	return f
}
