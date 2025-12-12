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
		t.Logf("%s\n", p.Where(err))
	}

	back := p.Format(n)

	t.Logf("%-26v -> %-26v  %v(%d)", text, back, n.Kind(), n.node.Arg())

	if back != exp {
		t.Errorf("root %v\n%#v", n.node, p)
	}
}
