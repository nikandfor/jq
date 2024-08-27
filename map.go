package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Map struct {
		Filter Filter

		arr []int
	}

	MapValues struct {
		Filter Filter

		arr []int
	}
)

func NewMap(f Filter) *Map             { return &Map{Filter: f} }
func NewMapValues(f Filter) *MapValues { return &MapValues{Filter: f} }

func (f *Map) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if next {
		return None, false, nil
	}

	res, f.arr, err = mapApplyTo(f.Filter, b, off, f.arr[:0], false)
	return res, false, err
}

func (f *MapValues) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if next {
		return None, false, nil
	}

	res, f.arr, err = mapApplyTo(f.Filter, b, off, f.arr[:0], true)
	return res, false, err
}

func mapApplyTo(f Filter, b *Buffer, off int, arr []int, values bool) (res int, _ []int, err error) {
	br := b.Reader()

	tag := br.Tag(off)
	if tag != cbor.Array && tag != cbor.Map {
		return off, arr, ErrType
	}

	val := 0
	if tag == cbor.Map {
		val = 1
	}

	arr = br.ArrayMap(off, arr)
	l := len(arr)

	for j := 0; j < l; j += 1 + val {
		if !values {
			arr, err = ApplyGetAll(f, b, arr[j+val], arr)
			if err != nil {
				return off, arr, err
			}

			continue
		}

		res, _, err = f.ApplyTo(b, arr[j+val], false)
		if err != nil {
			return off, arr, err
		}

		if res == None {
			continue
		}

		if tag == cbor.Map {
			arr = append(arr, arr[j])
		}

		arr = append(arr, res)
	}

	if !values {
		tag = cbor.Array
	}

	res = b.Writer().ArrayMap(tag, arr[l:])

	return res, arr, nil
}

func (f Map) String() string       { return fmt.Sprintf("map(%v)", f.Filter) }
func (f MapValues) String() string { return fmt.Sprintf("map_values(%v)", f.Filter) }
