package jqparser

import "testing"

func TestParser(t *testing.T) {
	var p Parser

	testParser(t, `.`, &p)
	testParser(t, `null`, &p)
	testParser(t, `true`, &p)
	testParser(t, `false`, &p)
	testParser(t, `.1`, &p)
	testParser(t, `"a"`, &p)
	testParser(t, `'a'`, &p)
	testParser(t, `.a`, &p)
	testParser(t, `$a`, &p)
	testParser(t, `1 as $a | $a`, &p)
	testParser(t, `.[4]`, &p)
	testParser(t, `.[1:2]`, &p)
	testParser(t, `.[:2]`, &p)
	testParser(t, `.[1:]`, &p)
	testParser(t, `.[:]`, &p)
	testParser(t, `.[]`, &p)
	testParser(t, `.[].qwe[][1]`, &p)
	testParser(t, `1 | 2`, &p)
	testParser(t, `1, 2`, &p)
	testParser(t, `1, 2 | 3, 4`, &p)
	testParser(t, `1, (2 | 3), 4`, &p)
	testParser(t, `1 + 2 * 3 - 4 % 5`, &p)
	testParser(t, `(1 + 2) * ((3 - 4) % 5)`, &p)
	testParser(t, `+-+-2`, &p)
	testParser(t, `-(1 - 2)`, &p)
	testParser(t, `[1, 2, 3]`, &p)
	testParser(t, `{a: "b", "c": 4, d, $e, (.f): .f}`, &p)
	testParser(t, `empty`, &p)
	testParser(t, `select(. > 3)`, &p)
	testParser(t, `.a.b.c("d")`, &p)
	testParser(t, `.("d")`, &p)
	testParser(t, `$a("d")`, &p)
	testParser(t, `$a.b.c("d")`, &p)
	testParser(t, `a("b").c(4)("d")`, &p)
	testParser(t, `qwe(. > 1, . < 2)`, &p)
	testParser(t, `qwe((. > 1, . < 2))`, &p)
	testParser(t, `if true then 1 end`, &p)
	testParser(t, `if true then 1 else 2 end`, &p)
	testParser(t, `true and false or false`, &p)
	testParser(t, `true and (false or false)`, &p)
	testParser2(t, `# comment
.`, `.`, &p)
	testParser2(t, `[
  1,
  # foo \
  2,
  # bar \\
  3,
  4, # baz \\\
  5, \
  6,
  7
  # comment \
    comment \
    comment
]`, `[1, 3, 4, 7]`, &p)

	/*
	   let f = x => x + 2
	   let y = f(3)

	   let y = (x => x + 2)(3)
	*/
}

func testParser(t *testing.T, text string, p *Parser) {
	t.Helper()
	testParser2(t, text, text, p)
}

func testParser2(t *testing.T, text, exp string, p *Parser) {
	t.Helper()

	n, err := p.Parse(text)
	if err != nil {
		t.Errorf("prog: %s\n\troot %v, err: %v", text, n, err)
		t.Logf("%s\n", err.(Error).Where())
	}

	var back string
	var arg int

	func() {
		//	defer func() {
		//		p := recover()
		//		if p != nil {
		//			t.Errorf("panic: %v", p)
		//		}
		//	}()

		back = p.Format(n)
		arg = n.node.Arg()
	}()

	t.Logf("%-26v -> %-26v  %v(%d)", text, back, n.Kind(), arg)

	if back != exp {
		t.Errorf("root %15v  %#v", n, n.node)

		for i, n := range p.nodes {
			t.Logf("node %3d  %10v  %#v  %6x", i, Node{n}, n, uint32(n))
		}
	}
}

func BenchmarkParserNew(b *testing.B) {
	b.ReportAllocs()

	text := `
	# \
		Showing worst case scenario \

	trim("contains escapes \n\"\\ \U0001F600 and non ASCII ñ") as $value |
	len($value) == 0x2A
	# let's introduce an error too \
	whatever
`

	var err error

	for range b.N {
		var p Parser

		_, err = p.Parse(text)
	}

	if err != nil {
		b.Errorf("error: %v", err)
	}
}

func BenchmarkParserNewErr(b *testing.B) {
	b.ReportAllocs()

	text := `
	# \
		Showing worst case scenario \

	trim("contains escapes \n\"\\ \U0001F600 and non ASCII ñ") as $value |
	len($value) == 0x2A
	# let's introduce an error too
	whatever
`

	var err error

	for range b.N {
		var p Parser

		_, err = p.Parse(text)
	}

	if err != nil {
		//	b.Errorf("error: %v", err)
	}
}

func BenchmarkParserReset(b *testing.B) {
	b.ReportAllocs()

	text := `
	# \
		Showing worst case scenario \

	trim("contains escapes \n\"\\ \U0001F600 and non ASCII ñ") as $value |
	len($value) == 0x2A
	# let's introduce an error too \
	whatever
`

	var err error
	var p Parser

	for range b.N {
		_, err = p.Parse(text)
	}

	if err != nil {
		b.Errorf("error: %v", err)
	}
}
