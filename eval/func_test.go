package eval

import "testing"

func TestParseFunc(t *testing.T) {
	testParseFunc(t, []byte("Func( a )"),
		FuncCall{
			Ident("Func"),
			List{Ident("a")},
		})

	testParseFunc(t, []byte("Func(a, b, c)"),
		FuncCall{
			Ident("Func"),
			List{
				Ident("a"),
				Ident("b"),
				Ident("c"),
			},
		})

	testParseFunc(t, []byte("Func()"), FuncCall{Ident("Func"), List(nil)})

	testParseFunc(t, []byte("Func"), FuncCall{Ident("Func"), None{}})
}

func testParseFunc(t *testing.T, data []byte, exp any) {
	testParse(t, FuncCallParser{
		Ident{},
		Optional{
			ListParser{
				Open: '(',
				Sep:  ',',
				Of: Spacer{
					Spaces: GoSpaces,
					Of:     Ident{},
				},
			},
		},
	}, data, exp)
}
