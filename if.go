package jq

import (
	"fmt"
)

type (
	If struct {
		Condition, Then, Else Filter

		cond, done   bool
		cnext, dnext bool
	}
)

func NewIf(c, t, e Filter) *If { return &If{Condition: c, Then: t, Else: e} }

func (f *If) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if !next {
		f.done = false
		f.cnext = false
		f.dnext = false
	}
	if f.done {
		return None, false, nil
	}

	if !f.dnext {
		cond := csel(f.Condition != nil, f.Condition, Filter(Dot{}))

		res, f.cnext, err = cond.ApplyTo(b, off, f.cnext)
		if err != nil {
			return off, false, err
		}

		f.cond = IsTrue(b, res)
	}

	var data Filter

	if f.cond {
		data = csel(f.Then != nil, f.Then, Filter(Dot{}))
	} else {
		data = csel(f.Else != nil, f.Else, Filter(Off(False)))
	}

	res, f.dnext, err = data.ApplyTo(b, off, f.dnext)
	if err != nil {
		return off, false, err
	}

	f.done = !(f.dnext || f.cnext)

	return res, !f.done, nil
}

func (f If) String() string {
	c := csel(f.Condition != nil, f.Condition, Filter(Dot{}))
	t := csel(f.Then != nil, f.Then, Filter(Dot{}))

	if f.Else == nil {
		return fmt.Sprintf("if %v then %v end", c, t)
	}

	return fmt.Sprintf("if %v then %v else %v end", c, t, f.Else)
}
