package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNum(t *testing.T) {
	testParseNum(t, []byte("1"), Int("1"))
	testParseNum(t, []byte("10"), Int("10"))
	testParseNum(t, []byte("012"), Int("012"))

	testParseNum(t, []byte("1.4"), Float("1.4"))
	testParseNum(t, []byte(".4"), Float(".4"))
	testParseNum(t, []byte("4."), Float("4."))
}

func testParseNum(t *testing.T, data []byte, exp any) {
	testParse(t, Num{}, data, exp)
}

func TestParseMath(t *testing.T) {
	testParseMath(t, []byte(" 1"), Int("1"))

	testParseMath(t, []byte("1 + 2"), BinOp{
		Op: Const("+"),
		L:  Int("1"),
		R:  Int("2"),
	})

	testParseMath(t, []byte("1 + 222 - 3"), BinOp{
		Op: Const("-"),
		L: BinOp{
			Op: Const("+"),
			L:  Int("1"),
			R:  Int("222"),
		},
		R: Int("3"),
	})

	testParseMath(t, []byte("a + 2"), BinOp{
		Op: Const("+"),
		L:  Ident("a"),
		R:  Int("2"),
	})

	testParseMath(t, []byte(" ( 1 + 2 ) * b"), BinOp{
		Op: Const("*"),
		L: BinOp{
			Op: Const("+"),
			L:  Int("1"),
			R:  Int("2"),
		},
		R: Ident("b"),
	})

	testParseMath(t, []byte(" b * ( 1 + 2 )"), BinOp{
		Op: Const("*"),
		L:  Ident("b"),
		R: BinOp{
			Op: Const("+"),
			L:  Int("1"),
			R:  Int("2"),
		},
	})
}

func testParseMath(t *testing.T, data []byte, exp any) {
	testParse(t, Math{}, data, exp)
}

func testParse(t *testing.T, root Parser, data []byte, exp any) {
	t.Run(string(data), func(t *testing.T) {
		x, i, err := root.Parse(data, 0)
		assert.NoError(t, err)
		assert.Equal(t, len(data), i)

		assert.Equal(t, exp, x, "want: %s\ngot:  %s", exp, x)
	})
}
