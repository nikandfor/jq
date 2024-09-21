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

		base Path
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
		f.rnext = false
		f.base = f.base[:0]
		f.path = f.path[:0]
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
		//	log.Printf("staring assign loop %v   %x [%x] %v", next, f.path, at, lnext)

		field, f.path, lnext, err = f.L.ApplyToGetPath(b, off, f.path[:0], lnext)
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

		for i := range f.base {
			if i == len(f.path) {
				break
			}

			f.path[i].Off = f.base[i].Off

			if f.base[i].Index != f.path[i].Index {
				break
			}
		}

		f.base = append(f.base[:0], f.path...)

		//	log.Printf("assign  %v#%v %v = %v %v", f.base, field, lnext, res, f.rnext)

		br := b.Reader()
		bw := b.Writer()

		for at := len(f.base) - 1; at >= 0; at-- {
			tag := br.Tag(f.base[at].Off)
			val := csel(tag == cbor.Map, 1, 0)

			index := f.base[at].Index
			index = csel(tag == cbor.Map, index*2, index)

			f.arr = br.ArrayMap(f.base[at].Off, f.arr[:0])

			if res != None {
				//	if f.arr[index+val] == res {
				//		continue
				//	}

				f.arr[index+val] = res
			} else {
				copy(f.arr[index:], f.arr[index+1+val:])
				f.arr = f.arr[:len(f.arr)-1-val]
			}

			res = bw.ArrayMap(tag, f.arr)
			f.base[at].Off = res
		}

		//	log.Printf("assigg  %v#%v", f.base, field)

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
