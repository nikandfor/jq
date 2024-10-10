package jq

import (
	"nikand.dev/go/cbor"
)

type (
	ToEntries struct {
		arr []Off

		Arrays bool
	}

	FromEntries struct {
		arr []Off
		obj []Off
	}
)

func NewToEntries() *ToEntries       { return &ToEntries{} }
func NewToEntriesArrays() *ToEntries { return &ToEntries{Arrays: true} }

func NewFromEntries() *FromEntries { return &FromEntries{} }

func NewWithEntries(f Filter) *Pipe {
	return NewPipe(NewToEntries(), NewMap(f), NewFromEntries())
}

func NewWithEntriesArrays(f Filter) *Pipe {
	return NewPipe(NewToEntriesArrays(), NewMap(f), NewFromEntries())
}

func (f *ToEntries) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()
	bw := b.Writer()

	if err = br.check(off, cbor.Map); err != nil {
		return None, false, fe(f, off, err)
	}

	f.arr = br.ArrayMap(off, f.arr[:0])

	if len(f.arr) == 0 {
		return EmptyArray, false, nil
	}

	var etag Tag
	var keyj, valj int
	l := len(f.arr)

	if f.Arrays {
		etag = cbor.Array
		keyj, valj = 0, 1

		f.arr = append(f.arr, Null, Null)
	} else {
		etag = cbor.Map
		keyj, valj = 1, 3
		key := bw.String("key")
		val := bw.String("value")

		f.arr = append(f.arr, key, Null, val, Null)
	}

	for j := 0; j < l; j += 2 {
		f.arr[l+keyj] = f.arr[j]
		f.arr[l+valj] = f.arr[j+1]

		f.arr = append(f.arr, bw.ArrayMap(etag, f.arr[l:l+valj+1]))
	}

	res = bw.Array(f.arr[l+valj+1:])

	return res, false, nil
}

func (f *FromEntries) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	br := b.Reader()

	if err = br.check(off, cbor.Array); err != nil {
		return None, false, fe(f, off, err)
	}

	f.arr = br.ArrayMap(off, f.arr[:0])

	if len(f.arr) == 0 {
		return Null, false, nil
	}

	l := len(f.arr)
	f.obj = f.obj[:0]

	for j := 0; j < l; j++ {
		if err = br.check2(f.arr[j], cbor.Map, cbor.Array); err != nil {
			return None, false, fe(f, f.arr[j], err)
		}

		f.arr = br.ArrayMap(f.arr[j], f.arr[:l])
		if len(f.arr) == 0 {
			continue
		}

		tag := br.Tag(f.arr[j])
		var key, val Off = None, Null

		if tag == cbor.Map {
			keyj := findKey(b, f.arr[l:], "key", "Key", "name", "Name")
			if keyj < 0 {
				return None, false, fe(f, f.arr[j], NewTypeError(br.Tag(Null), cbor.String, cbor.Bytes))
			}

			key = f.arr[l+keyj+1]
			valj := findKey(b, f.arr[l:], "value", "Value")
			if valj >= 0 {
				val = f.arr[l+valj+1]
			}
		} else if len(f.arr[l:]) == 0 {
			continue
		} else {
			key = f.arr[l]
			if len(f.arr) > 1 {
				val = f.arr[l+1]
			}
		}

		f.obj = append(f.obj, key, val)
	}

	res = b.Writer().Map(f.obj)

	return res, false, nil
}

func findKey(b *Buffer, obj []Off, keys ...string) (jj int) {
	br := b.Reader()
	jj = -1

	for _, key := range keys {
		for j := 0; j < len(obj); j += 2 {
			if tag := br.Tag(obj[j]); tag != cbor.String && tag != cbor.Bytes {
				continue
			}

			k := br.Bytes(obj[j])

			if string(k) == key {
				jj = j
			}
		}

		if jj >= 0 {
			return jj
		}
	}

	return -1
}

func (f *ToEntries) String() string   { return "to_entries" + csel(f.Arrays, "_arrays", "") }
func (f *FromEntries) String() string { return "from_entries" }
