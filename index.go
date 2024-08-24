package jq

import (
	"errors"
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Index struct {
		Path []any

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

func NewIndex(p ...any) *Index {
	return &Index{Path: p}
}

func (f *Index) ApplyTo(b *Buffer, off int, next bool) (res int, err error) {
	br := b.Reader()

	//	log.Printf("apply index %v to %x %v  state %v\n%s", f.Path, off, next, f.stack, DumpBuffer(b))

	//	defer func(off int) { log.Printf("index  %4x -> %4x  from %v", off, res, loc.Caller(1)) }(off)

	back := func(fi int) int {
		//	log.Printf("index %d go back  %v  %+v", fi, next, f.stack)
		if !next {
			return 0
		}
		if len(f.stack) == 0 {
			return -1
		}

		for ; fi >= 0; fi-- {
			st := f.stack[fi]
			if st.off < 0 {
				continue
			}
			if st.i != st.end {
				off = st.off
				return fi
			}

			f.stack[fi] = indexState{off: -1}
			f.arr = f.arr[:st.st]
		}

		if fi < 0 {
			off = None
		}

		return fi
	}

	fi := len(f.Path) - 1

back:
	for {
		fi = back(fi)
		//	log.Printf("index %d back  %4x", fi, off)
		if fi < 0 {
			return None, nil
		}

		for ; fi < len(f.Path); fi++ {
			//	log.Printf("index %d step  off %2x  key %v", fi, off, f.Path[fi])
			if off == Nil {
				continue back
				//	return Nil, nil
			}
			if off < 0 {
				panic(off)
			}

			tag := br.Tag(off)
			k := f.Path[fi]

			//	log.Printf("index tag %x  key %v", tag, k)

			switch k := k.(type) {
			case string:
				if tag != cbor.Map {
					return off, ErrType
				}

				//	q := off
				off = f.mapKey(b, off, k)

				//	log.Printf("index %d map  %x %v -> %x", fi, q, k, off)
			case int:
				//	q := off
				_, off = br.ArrayMapIndex(off, k)

				//	log.Printf("index %d arr  %x %v -> %x", fi, q, k, off)
			case Iter:
				if !next {
					//	log.Printf("index %d reset", fi)
					next = f.init()
				}

				if f.stack[fi].i == f.stack[fi].end {
					val := 0
					if tag == cbor.Map {
						val = 1
					}

					//	log.Printf("index %d iter init %x  val %v", fi, tag, val)

					st := len(f.arr)
					f.arr = br.ArrayMap(off, f.arr)

					f.stack[fi] = indexState{off: off, st: st, i: st, end: len(f.arr), val: val}
				}

				//	log.Printf("index %d iter  %+v  arr %v  path %+v", fi, f.stack, len(f.arr), f.Path)

				if f.stack[fi].i == f.stack[fi].end {
					continue back
				}

				val := f.stack[fi].val
				off = f.arr[f.stack[fi].i+val]
				f.stack[fi].i += 1 + val
			default:
				return off, ErrUnsupportedIndexKey
			}
		}

		break
	}

	return off, nil
}

func (f *Index) init() bool {
	f.arr = f.arr[:0]

	for cap(f.stack) < len(f.Path) {
		f.stack = append(f.stack[:cap(f.stack)], indexState{})
	}

	f.stack = f.stack[:len(f.Path)]

	for j := range f.stack {
		f.stack[j] = indexState{off: -1}
	}

	return true
}

func (f *Index) mapKey(b *Buffer, off int, key string) int {
	br := b.Reader()
	reset := len(f.arr)
	res := Nil

	f.arr = br.ArrayMap(off, f.arr)
	//	log.Printf("mapkey %x %q from %x", off, key, f.arr)

	for j := reset; j < len(f.arr); j += 2 {
		k := br.Bytes(f.arr[j])
		//	log.Printf("mapkey %3x %q -> %d %x  %q", off, key, j-reset, f.arr[j], k)
		if string(k) != key {
			continue
		}

		res = f.arr[j+1]
		break
	}

	f.arr = f.arr[:reset]

	return res
}

func (f Index) String() string { return fmt.Sprintf("Index{%v}", f.Path) }

func (s indexState) String() string {
	return fmt.Sprintf("{off %x, st %x, i %x, end %x, val %x}", s.off, s.st, s.i, s.end, s.val)
}
