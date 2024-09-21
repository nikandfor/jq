package jq

import "testing"

func TestSandwich(tb *testing.T) {
	s := NewSandwich(nil, nil, nil)

	for _, fn := range []SandwichFunc{s.ApplyGetOne, s.ApplyToOne, s.ApplyToAll} {
		w, err := fn(nil, nil)
		assertNoError(tb, err)
		if w != nil {
			tb.Errorf("expected nil")
		}
	}
}
