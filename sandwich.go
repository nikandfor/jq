package jq

import (
	"fmt"
)

type (
	Sandwich struct {
		Pre, Post Filter
		Filter    Filter
		Buffer    *Buffer
	}

	SandwichFunc = func(w, r []byte) ([]byte, error)
)

func NewSandwich(f, pre, post Filter) *Sandwich {
	return &Sandwich{
		Pre:    pre,
		Post:   post,
		Filter: f,
		Buffer: NewBuffer(nil),
	}
}

func (f *Sandwich) ApplyGetOne(w, r []byte) (_ []byte, err error) {
	f.Buffer.Reset(r)
	reset(f.Pre)
	reset(f.Post)

	var off Off

	if f.Pre != nil {
		off, _, err = f.Pre.ApplyTo(f.Buffer, 0, false)
		if err != nil {
			return w, fmt.Errorf("pre: %w", err)
		}
	}

	if f.Filter != nil {
		off, _, err = f.Filter.ApplyTo(f.Buffer, off, false)
		if err != nil {
			return w, fmt.Errorf("filter: %w", err)
		}
	}

	w, err = f.post(w, f.Buffer, off, false)
	if err != nil {
		return w, fmt.Errorf("post: %w", err)
	}

	return w, nil
}

func (f *Sandwich) ApplyToOne(w, r []byte) (_ []byte, err error) {
	f.Buffer.Reset(r)
	reset(f.Pre)
	reset(f.Post)

	var off Off

	if f.Pre != nil {
		off, _, err = f.Pre.ApplyTo(f.Buffer, 0, false)
		if err != nil {
			return w, fmt.Errorf("pre: %w", err)
		}
	}

	//	log.Printf("apply to one off %v\n%s", off, DumpBytes(len(f.Buffer.R), f.Buffer.W))

	w, err = f.filterAll(w, f.Buffer, off, false)
	if err != nil {
		return w, err
	}

	return w, nil
}

func (f *Sandwich) ApplyToAll(w, r []byte) (_ []byte, err error) {
	f.Buffer.Reset(r)
	reset(f.Pre)
	reset(f.Post)

	reset := f.Buffer.Writer().Off()
	pre, post := false, false

	for {
		var off Off
		f.Buffer.Writer().Reset(reset)

		if f.Pre != nil {
			off, pre, err = f.Pre.ApplyTo(f.Buffer, 0, pre)
			if err != nil {
				return w, fmt.Errorf("pre: %w", err)
			}
		}

		w, err = f.filterAll(w, f.Buffer, off, post)
		if err != nil {
			return w, err
		}

		if !pre {
			break
		}

		post = true
	}

	return w, nil
}

func (f *Sandwich) filterAll(w []byte, b *Buffer, off Off, post bool) (_ []byte, err error) {
	var res Off = None
	next := false

	//	log.Printf("filter all %v", off)

	for {
		if f.Filter != nil {
			res, next, err = f.Filter.ApplyTo(b, off, next)
			if err != nil {
				return w, fmt.Errorf("filter: %w", err)
			}
		}

		//	log.Printf("filter %v: %v -> %v  %v", f.Filter, off, res, next)

		w, err = f.post(w, b, res, post)
		if err != nil {
			return w, fmt.Errorf("post: %w", err)
		}

		if !next {
			break
		}

		post = true
	}

	return w, nil
}

func (f *Sandwich) post(w []byte, b *Buffer, off Off, next bool) (_ []byte, err error) {
	var res Off

	for f.Post != nil {
		res, next, err = f.Post.ApplyTo(b, off, next)
		if err != nil {
			return w, err
		}

		buf, st := b.Buf(res)
		w = append(w, buf[st:]...)

		if !next {
			break
		}
	}

	return w, nil
}

func reset(x any) {
	if r, ok := x.(interface{ Reset() }); ok {
		r.Reset()
	}
}
