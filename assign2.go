package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Assign struct {
		L FilterPath
		R Filter

		Relative bool

		rnext bool

		path Path
		arr  []Off
	}

	assignState struct {
		off        Off
		st, i, end int
	}
)

func NewAssign(l FilterPath, r Filter, rel bool) *Assign { return &Assign{L: l, R: r, Relative: rel} }

func (f *Assign) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if !next {
		f.path = append(f.path[:0], off)

		f.rnext = false

		f.arr = f.arr[:0]
	}

	var val Off

	for !f.Relative {
		val, more, err = f.R.ApplyTo(b, off, next)
		if err != nil {
			return off, false, err
		}
		if !more && val == None {
			return None, false, nil
		}
		if val == None {
			next = more
			continue
		}

		break
	}

	var lnext bool

	for {
		var field Off
		var at int
		//	log.Printf("staring assign loop %v   %x [%x] %v", next, f.path, at, lnext)

		field, f.path, at, lnext, err = f.L.ApplyToGetPath(b, f.path, at, lnext)
		if err != nil {
			return off, false, err
		}

		if field == None {
			if !lnext {
				break
			}

			continue
		}

		if f.Relative {
			res, more, err = f.R.ApplyTo(b, field, next)
			if err != nil {
				return off, false, err
			}
		} else {
			res = val
		}

		//	log.Printf("assign  %x:%x %v = %x %v", f.path[:at], field, lnext, res, f.rnext)

		br := b.Reader()
		bw := b.Writer()

		for at--; at >= 0; at-- {
			tag := br.Tag(f.path[at])
			val := csel(tag == cbor.Map, 1, 0)
			f.arr = br.ArrayMap(f.path[at], f.arr[:0])

			//	log.Printf("get fi %x  tag %x  %2x % x", at, tag, f.path[at], f.arr)

			for j := 0; j < len(f.arr); j += 1 + val {
				if f.arr[j+val] != field {
					continue
				}

				if res != None {
					f.arr[j+val] = res
					break
				}

				copy(f.arr[j:], f.arr[j+1+val:])
				f.arr = f.arr[:len(f.arr)-1-val]

				break
			}

			field = f.path[at]
			res = bw.ArrayMap(tag, f.arr)
			f.path[at] = res

			//	log.Printf("set fi %x  tag %x  %2x % x", at, tag, f.path[at], f.arr)
		}

		if !lnext {
			break
		}
	}

	return res, more, nil
}

func (f *Assign) String() string {
	op := "="
	if f.Relative {
		op = "|="
	}

	return fmt.Sprintf("%v %s %v", f.L, op, f.R)
}
