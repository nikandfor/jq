package jq

import "testing"

func TestEntries(tb *testing.T) {
	b := NewBuffer()
	r := b.appendVal(obj{"a", 1, "b", "c", "d", arr{}, "e", obj{}})

	key := b.appendVal("key")
	val := b.appendVal("value")

	testOne(tb, NewToEntries(), b, r,
		arr{
			obj{key, "a", val, 1},
			obj{key, "b", val, "c"},
			obj{key, "d", val, arr{}},
			obj{key, "e", val, obj{}},
		},
	)

	testOne(tb, NewToEntriesArrays(), b, r,
		arr{
			arr{"a", 1},
			arr{"b", "c"},
			arr{"d", arr{}},
			arr{"e", obj{}},
		},
	)

	testOne(tb, NewPipe(NewToEntries(), NewFromEntries()), b, r, r)
	testOne(tb, NewPipe(NewToEntriesArrays(), NewFromEntries()), b, r, r)

	testOne(tb, NewWithEntriesArrays(NewComma(
		NewArray(NewPlus(Index(0), NewLiteral("_first")), Index(1)),
		NewArray(NewPlus(Index(0), NewLiteral("_second")), Zero),
	)), b, r,
		obj{
			"a_first", 1,
			"a_second", 0,
			"b_first", "c",
			"b_second", 0,
			"d_first",
			arr{},
			"d_second", 0,
			"e_first",
			obj{},
			"e_second", 0,
		},
	)
}
