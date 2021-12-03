package scanner

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNext(t *testing.T) {
	var s Scanner

	in := `a = 3; "qwe"`

	r := strings.NewReader(in)

	s.Reset(r)

	c := s.Next()
	for c >= 0 {
		c = s.Next()
	}

	assert.True(t, c == Err && s.Err() == io.EOF, "tk: %v  err %v", c, s.Err())

	assert.Equal(t, in, string(s.b[s.st:s.end]))
}

func TestScan(t *testing.T) {
	testScan(t, `abc = 321; "qwe"`, "abc", "=", "321", ";", `"qwe"`)
}

func testScan(t *testing.T, q string, exp ...string) {
	var s Scanner

	r := strings.NewReader(q)

	s.Reset(r)

	var res []string

	var tk Token
	for {
		tk = s.Scan()

		if tk == Err {
			break
		}

		res = append(res, string(s.TokenText()))
	}

	assert.True(t, tk == Err && s.Err() == io.EOF, "tk: %v  err %v", tk, s.Err())

	assert.Equal(t, exp, res, "%v", q)
}
