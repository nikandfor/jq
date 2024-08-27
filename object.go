package jq

import (
	"fmt"
	"strings"
)

type (
	Object struct {
		Keys []ObjectKey

		arr   []int
		stack []objectState
	}

	ObjectKey struct {
		Key   Filter
		Value Filter
	}

	ObjectCopyKey string

	objectState struct {
		next bool
	}
)

func NewObject(kvs ...any) *Object {
	var e Encoder
	var b []byte

	obj := make([]ObjectKey, 0, len(kvs)/2)

	for i := 0; i < len(kvs); {
		var key Filter

		if k, ok := kvs[i].(string); ok {
			st := len(b)
			b = e.AppendString(b, k)

			key = Literal(b[st:])
		} else if s, ok := kvs[i].(ObjectCopyKey); ok {
			st := len(b)
			b = e.AppendString(b, string(s))

			key = Literal(b[st:])

			obj = append(obj, ObjectKey{
				Key:   key,
				Value: NewIndex(string(s)),
			})

			i++
			continue
		} else if k, ok := kvs[i].(Filter); ok {
			key = k
		} else {
			panic(kvs[i])
		}

		obj = append(obj, ObjectKey{
			Key:   key,
			Value: kvs[i+1].(Filter),
		})

		i += 2
	}

	return &Object{Keys: obj}
}

func (f *Object) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	bw := b.Writer()

	//	defer func(off int) { log.Printf("object.Apply %x %v  =>  %x %v", off, next, res, err) }(off)

	reset := bw.Len()
	defer bw.ResetIfErr(reset, &err)

	if !next {
		f.init()
	}

	back := func(fi int) int {
		if !next {
			return 0
		}

		for ; fi >= 0; fi-- {
			if f.arr[fi] >= 0 && f.stack[fi].next {
				break
			}

			f.stack[fi] = objectState{}
		}

		return fi
	}

	fi := 2*len(f.Keys) - 1

back:
	for {
		fi = back(fi)
		//	log.Printf("back %d %v  %x %v", fi, next, f.arr, f.stack)
		if fi < 0 {
			return None, false, nil
		}

		next = true

		for ; fi < 2*len(f.Keys); fi++ {
			st := f.stack[fi]
			kv := f.Keys[fi/2]

			var subf Filter

			if fi%2 == 0 {
				subf = kv.Key
			} else {
				subf = kv.Value
			}

			f.arr[fi], f.stack[fi].next, err = subf.ApplyTo(b, off, st.next)
			//	log.Printf("obj iter %d: %3x %v => %3x  (bsize %3x)", fi, off, st.next, f.arr[fi], bw.Offset())
			if err != nil {
				return None, false, err
			}

			if f.arr[fi] == None /*|| f.arr[fi] == Null && fi%2 == 0*/ {
				continue back
			}
		}

		break
	}

	off = bw.Map(f.arr)

	more = back(2*len(f.Keys)-1) >= 0

	return off, more, nil
}

func (f *Object) init() bool {
	for cap(f.stack) < 2*len(f.Keys) {
		f.stack = append(f.stack[:cap(f.stack)], objectState{})
	}

	f.stack = f.stack[:2*len(f.Keys)]

	if cap(f.arr) < len(f.stack) {
		f.arr = make([]int, cap(f.stack))
	}

	f.arr = f.arr[:len(f.stack)]

	return true
}

func (f Object) String() string {
	var b strings.Builder

	//	b.WriteString("Object{")
	b.WriteString("{")

	for i, kv := range f.Keys {
		if i != 0 {
			b.WriteString(", ")
		}

		switch k := kv.Key.(type) {
		case Literal:
			fmt.Fprintf(&b, "%v", k)
		default:
			fmt.Fprintf(&b, "(%v)", kv.Key)
		}

		fmt.Fprintf(&b, ": %v", kv.Value)
	}

	b.WriteString("}")

	return b.String()
}
