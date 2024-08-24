package jq

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

	objectState struct {
		next bool
	}
)

func NewObject(kvs ...any) *Object {
	var e Encoder
	var b []byte

	obj := make([]ObjectKey, len(kvs)/2)

	for i := 0; i < len(kvs); i += 2 {
		var key Filter

		if k, ok := kvs[i].(string); ok {
			st := len(b)
			b = e.AppendString(b, k)

			key = Literal(b[st:])
		} else if k, ok := kvs[i].(Filter); ok {
			key = k
		} else {
			panic(kvs[i])
		}

		obj[i/2] = ObjectKey{
			Key:   key,
			Value: kvs[i+1].(Filter),
		}
	}

	return &Object{Keys: obj}
}

func (f *Object) ApplyTo(b *Buffer, off int, next bool) (res int, err error) {
	bw := b.Writer()

	//	defer func(off int) { log.Printf("object.Apply %x %v  =>  %x %v", off, next, res, err) }(off)

	reset := bw.Offset()
	defer bw.ResetIfErr(reset, &err)

	if !next {
		f.init()
	}

	back := func(fi int) int {
		if !next {
			return 0
		}

		for ; fi >= 0; fi-- {
			if f.arr[fi] >= 0 {
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
			return None, nil
		}

		for ; fi < 2*len(f.Keys); fi++ {
			st := f.stack[fi]
			kv := f.Keys[fi/2]

			var subf Filter

			if fi%2 == 0 {
				subf = kv.Key
			} else {
				subf = kv.Value
			}

			f.arr[fi], err = subf.ApplyTo(b, off, st.next)
			//	log.Printf("obj iter %d: %3x %v => %3x  (bsize %3x)", fi, off, st.next, f.arr[fi], bw.Offset())
			f.stack[fi].next = true
			if err != nil {
				return None, err
			}

			if f.arr[fi] == None || f.arr[fi] == Nil && fi%2 == 0 {
				continue back
			}
		}

		break
	}

	off = bw.Map(f.arr)

	//	log.Printf("object %x  %x", off, f.arr)
	//	d := Dumper{Writer: os.Stderr}
	//	d.ApplyTo(b, 0, false)

	return off, nil
}

func (f *Object) init() bool {
	for cap(f.stack) < len(f.Keys)*2 {
		f.stack = append(f.stack[:cap(f.stack)], objectState{})
	}

	f.stack = f.stack[:2*len(f.Keys)]

	if cap(f.arr) < len(f.stack) {
		f.arr = make([]int, cap(f.stack))
	}

	f.arr = f.arr[:len(f.stack)]

	return true
}
