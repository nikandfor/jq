package jq

import "testing"

func TestSandwich(tb *testing.T) {
	s := NewSandwich(nil, nil)

	for _, fn := range []SandwichFunc{s.ProcessGetOne, s.ProcessOne, s.ProcessAll} {
		w, err := fn(nil, nil, nil)
		assertNoError(tb, err)
		if w != nil {
			tb.Errorf("expected nil")
		}
	}
}
