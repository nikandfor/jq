package jq

import (
	"fmt"
)

type (
	DataDecoder interface {
		Decode(b *Buffer, r []byte, st int) (Off, int, error)
	}

	DataEncoder interface {
		Encode(w []byte, b *Buffer, off Off) ([]byte, error)
	}

	Sandwich struct {
		Decoder DataDecoder
		Encoder DataEncoder

		Buffer *Buffer
	}

	SandwichFunc = func(f Filter, w, r []byte) ([]byte, error)
)

func NewSandwich(d DataDecoder, e DataEncoder) *Sandwich {
	return &Sandwich{
		Decoder: d,
		Encoder: e,
		Buffer:  NewBuffer(),
	}
}

func (s *Sandwich) Reset() {
	reset(s.Decoder)
	reset(s.Encoder)
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

	w, err = s.ApplyEncodeAll(f, w, off)
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

		off, i, err = s.Decode(r, i)
		if err != nil {
			return w, fmt.Errorf("decode: %w", err)
		}

		if off == None {
			return w, nil
		}

		w, err = s.ApplyEncodeAll(f, w, off)
		if err != nil {
			return w, err
		}
	}

	return w, nil
}

func (s *Sandwich) DecodeApply(f Filter, r []byte, st int) (res Off, more bool, i int, err error) {
	res, i, err = s.Decode(r, st)
	if err != nil {
		return None, false, i, fmt.Errorf("decode: %w", err)
	}

	res, more, err = s.ApplyTo(f, res, false)
	if err != nil {
		return None, false, i, fmt.Errorf("apply: %w", err)
	}

	return
}

func (s *Sandwich) ApplyEncodeOne(f Filter, w []byte, off Off) (_ []byte, err error) {
	res, _, err := s.ApplyTo(f, off, false)
	if err != nil {
		return w, fmt.Errorf("apply: %w", err)
	}

	//	log.Printf("apply %v: %v -> %v  %v", f, off, res, next)

	w, err = s.Encode(w, res)
	if err != nil {
		return w, fmt.Errorf("encode: %w", err)
	}

	return w, nil
}

func (s *Sandwich) ApplyEncodeAll(f Filter, w []byte, off Off) (_ []byte, err error) {
	var res Off = None
	next := false

	for {
		res, next, err = s.ApplyTo(f, off, next)
		if err != nil {
			return w, fmt.Errorf("apply: %w", err)
		}

		//	log.Printf("apply %v: %v -> %v  %v", f, off, res, next)

		w, err = s.Encode(w, res)
		if err != nil {
			return w, fmt.Errorf("encode: %w", err)
		}

		if !next {
			break
		}
	}

	return w, nil
}

func (s *Sandwich) ApplyTo(f Filter, off Off, next bool) (Off, bool, error) {
	if f == nil {
		return off, next, nil
	}

	return f.ApplyTo(s.Buffer, off, next)
}

func (s *Sandwich) Decode(r []byte, st int) (Off, int, error) {
	if s.Decoder == nil {
		return None, st, nil
	}

	return s.Decoder.Decode(s.Buffer, r, st)
}

func (s *Sandwich) Encode(w []byte, off Off) ([]byte, error) {
	if s.Encoder == nil || off == None {
		return w, nil
	}

	return s.Encoder.Encode(w, s.Buffer, off)
}

func reset(x any) {
	if r, ok := x.(interface{ Reset() }); ok {
		r.Reset()
	}
}
