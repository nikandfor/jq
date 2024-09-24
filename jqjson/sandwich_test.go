package jqjson

import (
	"bytes"
	"testing"

	"nikand.dev/go/jq"
)

func TestSandwichNil(tb *testing.T) {
	var w []byte

	d := NewDecoder()
	e := NewEncoder()

	s := jq.NewSandwich(d, e)

	for _, fn := range []jq.SandwichFunc{s.ProcessGetOne, s.ProcessOne, s.ProcessAll} {
		wst := len(w)
		w, err := fn(nil, w, []byte(``))
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

	d := NewDecoder()
	e := NewEncoder()
	e.Separator = []byte{' '}

	f := jq.NewQuery("a", jq.Iter{})
	s := jq.NewSandwich(d, e)

	r := []byte(`{"a":[1,2]}{"a":[]}{"a":[3]}`)

	wst := len(w)
	w, err := s.ProcessGetOne(f, w, r)
	if err != nil {
		tb.Errorf("apply get one: %v", err)
		return
	}

	exp := []byte(`1`)
	if !bytes.Equal(exp, w[wst:]) {
		tb.Errorf("\nexp (%s)\ngot (%s)", exp, w[wst:])
	}

	wst = len(w)
	w, err = s.ProcessOne(f, w, r)
	if err != nil {
		tb.Errorf("apply to one: %v", err)
		return
	}

	exp = []byte(`1 2`)
	if !bytes.Equal(exp, w[wst:]) {
		tb.Errorf("\nexp (%s)\ngot (%s)", exp, w[wst:])
	}

	wst = len(w)
	w, err = s.ProcessAll(f, w, r)
	if err != nil {
		tb.Errorf("apply to all: %v", err)
		return
	}

	exp = []byte(`1 2 3`)
	if !bytes.Equal(exp, w[wst:]) {
		tb.Errorf("\nexp (%s)\ngot (%s)", exp, w[wst:])
	}
}

func TestSandwichElements(tb *testing.T) {
	var w []byte

	d := NewDecoder()
	e := NewEncoder()
	e.Separator = []byte{',', '\n'}

	f := jq.NewQuery("a", jq.Iter{})
	s := jq.NewSandwich(d, e)

	r := []byte(`{"a":[1,2,3,4],"b":"c"}`)

	wst := len(w)
	w, err := s.ProcessAll(f, w, r)
	if err != nil {
		tb.Errorf("apply to all: %v", err)
		return
	}

	exp := []byte("1,\n2,\n3,\n4")
	if !bytes.Equal(exp, w[wst:]) {
		tb.Errorf("\nexp (%s)\ngot (%s)", exp, w[wst:])
	}
}
