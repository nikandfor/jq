package jq

import "fmt"

type (
	Any struct {
		Generator Filter
		Condition Filter
	}

	All struct {
		Generator Filter
		Condition Filter
	}
)

func NewAny(gen, cond Filter) *Any { return &Any{Generator: gen, Condition: cond} }
func NewAll(gen, cond Filter) *All { return &All{Generator: gen, Condition: cond} }

func (f *Any) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	res, err = anyAllApplyTo(b, off, f.Generator, f.Condition, 0)
	return res, false, err
}

func (f *All) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	res, err = anyAllApplyTo(b, off, f.Generator, f.Condition, 1)
	return res, false, err
}

func anyAllApplyTo(b *Buffer, off Off, gen, cond Filter, flip Off) (res Off, err error) {
	res = False ^ flip

	b.arr, err = iterOrGen(b, off, gen, func(right Off) (bool, error) {
		ok, err := condition(b, right, cond, flip)
		if err != nil {
			return false, err
		}

		if (flip == 0) == (ok == True) {
			res = ok
			return false, nil
		}

		return true, nil
	}, b.arr)
	if err != nil {
		return off, err
	}

	return res, nil
}

func condition(b *Buffer, off Off, cond Filter, flip Off) (res Off, err error) {
	if cond == nil {
		return ToBool(b, off), nil
	}

	next := false

	for {
		res, next, err = cond.ApplyTo(b, off, next)
		if err != nil {
			return off, err
		}
		if res != None {
			if IsTrue(b, res) == (flip == 0) {
				return True ^ flip, nil
			}
		}
		if !next {
			break
		}
	}

	return False ^ flip, nil
}

func (f Any) String() string {
	if f.Generator != nil && f.Condition != nil {
		return fmt.Sprintf("any(%v; %v)", f.Generator, f.Condition)
	}
	if f.Condition != nil {
		return fmt.Sprintf("any(%v)", f.Condition)
	}

	return "any"
}

func (f All) String() string {
	if f.Generator != nil && f.Condition != nil {
		return fmt.Sprintf("all(%v; %v)", f.Generator, f.Condition)
	}
	if f.Condition != nil {
		return fmt.Sprintf("all(%v)", f.Condition)
	}

	return "all"
}
