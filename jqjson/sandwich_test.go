package jqjson

import (
	"bytes"
	"testing"

	"nikand.dev/go/jq"
)

func TestSandwichNil(tb *testing.T) {
	var w []byte

	d := NewRawDecoder()
	e := NewRawEncoder()

	s := jq.NewSandwich(nil, d, e)

	for _, fn := range []jq.SandwichFunc{s.ApplyGetOne, s.ApplyToOne, s.ApplyToAll} {
		wst := len(w)
		w, err := fn(w, []byte(``))
		if err != nil {
			tb.Errorf("apply: %v", err)
			break
		}
		if !bytes.Equal(nil, w) {
			tb.Errorf("expected (%s)\ngot (%s)", []byte{}, w[wst:])
			break
		}
	}
}

func TestSandwich(tb *testing.T) {
	var w []byte

	d := NewRawDecoder()
	e := NewRawEncoder()
	e.Separator = []byte{' '}

	s := jq.NewSandwich(jq.NewQuery("a", jq.Iter{}), d, e)

	r := []byte(`{"a":[1,2]}{"a":[]}{"a":[3]}`)

	wst := len(w)
	w, err := s.ApplyGetOne(w, r)
	if err != nil {
		tb.Errorf("apply get one: %v", err)
		return
	}

	exp := []byte(`1`)
	if !bytes.Equal(exp, w[wst:]) {
		tb.Errorf("\nexp (%s)\ngot (%s)", exp, w[wst:])
	}

	wst = len(w)
	w, err = s.ApplyToOne(w, r)
	if err != nil {
		tb.Errorf("apply to one: %v", err)
		return
	}

	exp = []byte(`1 2`)
	if !bytes.Equal(exp, w[wst:]) {
		tb.Errorf("\nexp (%s)\ngot (%s)", exp, w[wst:])
	}

	wst = len(w)
	w, err = s.ApplyToAll(w, r)
	if err != nil {
		tb.Errorf("apply to all: %v", err)
		return
	}

	exp = []byte(`1 2 3`)
	if !bytes.Equal(exp, w[wst:]) {
		tb.Errorf("\nexp (%s)\ngot (%s)", exp, w[wst:])
	}
}
