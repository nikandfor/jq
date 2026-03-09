package jq

import "testing"

func TestHas(tb *testing.T) {
	b := NewBuffer()

	testSimple(tb, HasIndex(0), b, arr{}, false)
	testSimple(tb, HasIndex(1), b, arr{0}, false)
	testSimple(tb, HasIndex(-2), b, arr{0}, false)

	testSimple(tb, NewHas(0), b, arr{0}, true)
	testSimple(tb, NewHas(-1), b, arr{0}, true)

	//

	testSimple(tb, HasKey("a"), b, obj{}, false)
	testSimple(tb, HasKey("a"), b, obj{"b", 1}, false)

	testSimple(tb, NewHas("a"), b, obj{"a", 1}, true)

	//

	testSimple(tb, NewHas(NewLiteral(0)), b, arr{}, false)
	testSimple(tb, NewHas(NewLiteral(1)), b, arr{0}, false)
	testSimple(tb, NewHas(NewLiteral(-2)), b, arr{0}, false)

	testSimple(tb, NewHas(NewLiteral(0)), b, arr{0}, true)
	testSimple(tb, NewHas(NewLiteral(-1)), b, arr{0}, true)

	//

	testSimple(tb, NewHas(NewLiteral("a")), b, obj{}, false)
	testSimple(tb, NewHas(NewLiteral("a")), b, obj{"b", 1}, false)

	testSimple(tb, NewHas(NewLiteral("a")), b, obj{"a", 1}, true)
}
