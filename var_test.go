package jq

import "testing"

func TestBindVar(tb *testing.T) {
	b := NewBuffer()

	b.Bind("var", Zero)

	av := b.appendVal(10)
	bv := b.appendVal(20)
	obj := b.appendVal(obj{"a", av, "b", bv})

	f := NewPipe(Bind("var_a", Key("a")), Bind("var_b", Key("b")), One, Var("var_a"))

	testOne(tb, f, b, obj, av)

	assertDeepEqual(tb, []Variable{{Name: "var", Value: Zero}}, b.Vars)
}

func TestBindVarIter(tb *testing.T) {
	b := NewBuffer()

	b.Bind("var", Zero)

	a := b.appendVal(arr{1, 2, 3})

	testIter(tb, NewPipe(Bind("v", NewIter()), Bind("v1", EmptyString), Var("v")), b, a, []any{1, 2, 3})

	assertDeepEqual(tb, []Variable{{Name: "var", Value: Zero}}, b.Vars)
}
