package jq

import (
	"fmt"

	"nikand.dev/go/cbor"
)

type (
	Add struct {
		Generator Filter
	}
)

func NewAdd() Add              { return Add{} }
func NewAddGen(gen Filter) Add { return Add{gen} }

func (f Add) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	left := Null

	b.arr, err = iterOrGen(b, off, f.Generator, func(right Off) (bool, error) {
		left, err = plus(b, left, right)
		return true, err
	}, b.arr)
	if err != nil {
		return off, false, err
	}

	return left, false, nil
}

func iterOrGen(b *Buffer, off Off, gen Filter, f func(off Off) (bool, error), arr []Off) (_ []Off, err error) {
	if gen != nil {
		return arr, generator(b, off, gen, f)
	}

	br := b.Reader()
	off = br.UnderAllLabels(off)

	tag := br.Tag(off)
	if tag != TagArray && tag != TagMap {
		return arr, NewTypeError(tag, TagArray, TagMap)
	}

	arrbase := len(b.arr)
	defer func() { b.arr = b.arr[:arrbase] }()

	b.arr = br.ArrayMap(off, b.arr)
	val := csel(tag == cbor.Map, 1, 0)

	for i := arrbase; i < len(b.arr); i += 1 + val {
		ok, err := f(b.arr[i+val])
		if err != nil || !ok {
			return arr, err
		}
	}

	return arr, nil
}

func generator(b *Buffer, off Off, gen Filter, f func(off Off) (bool, error)) (err error) {
	next := false
	res := None

	for {
		res, next, err = gen.ApplyTo(b, off, next)
		if err != nil {
			return err
		}
		if res != None {
			ok, err := f(res)
			if err != nil || !ok {
				return err
			}
		}
		if !next {
			break
		}
	}

	return nil
}

func AddMaps(b *Buffer, objs ...Off) (Off, error) {
	if len(objs) == 0 {
		return Null, nil
	}

	arrbase := len(b.arr)
	defer func() { b.arr = b.arr[:arrbase] }()

	var err error
	end := arrbase

	for _, off := range objs {
		b.arr, end, err = addMaps(b, b.arr, arrbase, end, off)
		if err != nil {
			return None, err
		}
	}

	return b.Writer().Map(b.arr[arrbase:end]), nil
}

func addMaps(b *Buffer, arr []Off, base, end int, off Off) ([]Off, int, error) {
	arr = b.Reader().ArrayMap(off, arr[:end])

	if len(arr) == end {
		return arr, len(arr), nil
	}

out:
	for i := end; i < len(arr); i += 2 {
		for j := base; j < end; j += 2 {
			if b.Equal(arr[j], arr[i]) {
				arr[j+1] = arr[i+1]
				continue out
			}
		}

		if end != i {
			arr[end], arr[end+1] = arr[i], arr[i+1]
		}

		end += 2
	}

	return arr, end, nil
}

func (f Add) String() string {
	if f.Generator == nil {
		return "add"
	}

	return fmt.Sprintf("add(%v)", f.Generator)
}
