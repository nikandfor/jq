package jq

import (
	"fmt"
)

type (
	FormatDecoder interface {
		Decode(b *Buffer, r []byte, st int) (Off, int, error)
	}

	FormatEncoder interface {
		Encode(w []byte, b *Buffer, off Off) ([]byte, error)
	}

	Sandwich struct {
		Decoder FormatDecoder
		Encoder FormatEncoder

		Buffer *Buffer
	}

	SandwichFunc = func(f Filter, w, r []byte) ([]byte, error)
)

func NewSandwich(d FormatDecoder, e FormatEncoder) *Sandwich {
	return &Sandwich{
		Decoder: d,
		Encoder: e,
		Buffer:  NewBuffer(),
	}
}

func (s *Sandwich) Reset(d FormatDecoder, e FormatEncoder) {
	s.Decoder, s.Encoder = d, e
	s.Buffer.Reset()
}

func (s *Sandwich) ProcessGetOne(f Filter, w, r []byte) (_ []byte, err error) {
	reset(s.Decoder)
	reset(s.Encoder)

	var off Off

	if s.Decoder != nil {
		off, _, err = s.Decoder.Decode(s.Buffer, r, 0)
		if err != nil {
			return w, fmt.Errorf("decode: %w", err)
		}

		if off == None {
			return w, nil
		}
	}

	if f != nil {
		off, _, err = f.ApplyTo(s.Buffer, off, false)
		if err != nil {
			return w, fmt.Errorf("filter: %w", err)
		}
	}

	if s.Encoder != nil && off != None {
		w, err = s.Encoder.Encode(w, s.Buffer, off)
		if err != nil {
			return w, fmt.Errorf("encode: %w", err)
		}
	}

	return w, nil
}

func (s *Sandwich) ProcessOne(f Filter, w, r []byte) (_ []byte, err error) {
	reset(s.Decoder)
	reset(s.Encoder)

	var off Off

	if s.Decoder != nil {
		off, _, err = s.Decoder.Decode(s.Buffer, r, 0)
		if err != nil {
			return w, fmt.Errorf("decode: %w", err)
		}

		if off == None {
			return w, nil
		}
	}

	w, err = s.ApplyAll(f, w, off)
	if err != nil {
		return w, err
	}

	return w, nil
}

func (s *Sandwich) ProcessAll(f Filter, w, r []byte) (_ []byte, err error) {
	reset(s.Decoder)
	reset(s.Encoder)

	i := 0

	for i < len(r) {
		off := Null

		if s.Decoder != nil {
			off, i, err = s.Decoder.Decode(s.Buffer, r, i)
			if err != nil {
				return w, fmt.Errorf("decode: %w", err)
			}

			if off == None {
				return w, nil
			}
		}

		w, err = s.ApplyAll(f, w, off)
		if err != nil {
			return w, err
		}
	}

	return w, nil
}

func (s *Sandwich) ApplyOne(f Filter, w []byte, off Off) (_ []byte, err error) {
	f = csel(f != nil, f, (Filter)(Dot{}))

	res, _, err := f.ApplyTo(s.Buffer, off, false)
	if err != nil {
		return w, fmt.Errorf("apply: %w", err)
	}

	//	log.Printf("apply %v: %v -> %v  %v", f, off, res, next)

	if s.Encoder != nil && res != None {
		w, err = s.Encoder.Encode(w, s.Buffer, res)
		if err != nil {
			return w, fmt.Errorf("encode: %w", err)
		}
	}

	return w, nil
}

func (s *Sandwich) ApplyAll(f Filter, w []byte, off Off) (_ []byte, err error) {
	var res Off = None
	next := false

	f = csel(f != nil, f, (Filter)(Dot{}))

	for {
		res, next, err = f.ApplyTo(s.Buffer, off, next)
		if err != nil {
			return w, fmt.Errorf("apply: %w", err)
		}

		//	log.Printf("apply %v: %v -> %v  %v", f, off, res, next)

		if s.Encoder != nil && res != None {
			w, err = s.Encoder.Encode(w, s.Buffer, res)
			if err != nil {
				return w, fmt.Errorf("encode: %w", err)
			}
		}

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
