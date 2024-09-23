package jq

import (
	"fmt"
)

type (
	Sandwich struct {
		Pre, Post Filter
		Buffer    *Buffer
	}

	SandwichFunc = func(f Filter, w, r []byte) ([]byte, error)
)

func NewSandwich(pre, post Filter) *Sandwich {
	return &Sandwich{
		Pre:    pre,
		Post:   post,
		Buffer: NewBuffer(nil),
	}
}

func (s *Sandwich) ProcessGetOne(f Filter, w, r []byte) (_ []byte, err error) {
	s.Buffer.Reset(r)
	reset(s.Pre)
	reset(s.Post)

	var off Off

	if s.Pre != nil {
		off, _, err = s.Pre.ApplyTo(s.Buffer, 0, false)
		if err != nil {
			return w, fmt.Errorf("pre: %w", err)
		}
	}

	if f != nil {
		off, _, err = f.ApplyTo(s.Buffer, off, false)
		if err != nil {
			return w, fmt.Errorf("filter: %w", err)
		}
	}

	w, err = s.post(w, s.Buffer, off, false)
	if err != nil {
		return w, fmt.Errorf("post: %w", err)
	}

	return w, nil
}

func (s *Sandwich) ProcessOne(f Filter, w, r []byte) (_ []byte, err error) {
	s.Buffer.Reset(r)
	reset(s.Pre)
	reset(s.Post)

	var off Off

	if s.Pre != nil {
		off, _, err = s.Pre.ApplyTo(s.Buffer, 0, false)
		if err != nil {
			return w, fmt.Errorf("pre: %w", err)
		}
	}

	//	log.Printf("apply to one off %v\n%s", off, DumpBytes(len(s.Buffer.R), s.Buffer.W))

	w, err = s.filterAll(f, w, s.Buffer, off, false)
	if err != nil {
		return w, err
	}

	return w, nil
}

func (s *Sandwich) ProcessAll(f Filter, w, r []byte) (_ []byte, err error) {
	s.Buffer.Reset(r)
	reset(s.Pre)
	reset(s.Post)

	reset := s.Buffer.Writer().Off()
	pre, post := false, false

	for {
		var off Off
		s.Buffer.Writer().Reset(reset)

		if s.Pre != nil {
			off, pre, err = s.Pre.ApplyTo(s.Buffer, 0, pre)
			if err != nil {
				return w, fmt.Errorf("pre: %w", err)
			}
		}

		w, err = s.filterAll(f, w, s.Buffer, off, post)
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

func (s *Sandwich) filterAll(f Filter, w []byte, b *Buffer, off Off, post bool) (_ []byte, err error) {
	var res Off = None
	next := false

	f = csel(f != nil, f, (Filter)(Dot{}))

	//	log.Printf("filter all %v", off)

	for {
		res, next, err = f.ApplyTo(b, off, next)
		if err != nil {
			return w, fmt.Errorf("filter: %w", err)
		}

		//	log.Printf("filter %v: %v -> %v  %v", f, off, res, next)

		w, err = s.post(w, b, res, post)
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

func (s *Sandwich) post(w []byte, b *Buffer, off Off, next bool) (_ []byte, err error) {
	var res Off

	for s.Post != nil {
		res, next, err = s.Post.ApplyTo(b, off, next)
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
