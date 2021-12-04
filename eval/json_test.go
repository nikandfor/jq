package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	testJSON(t, `null`, Null{})
	testJSON(t, `true`, true)
	testJSON(t, `false`, false)
	testJSON(t, `"string"`, String(`"string"`))
	testJSON(t, `[1, 2]`, List{Int("1"), Int("2")})
	testJSON(t, `{"qwe": null}`, List{KVPair{K: String(`"qwe"`), V: Null{}}})
}

func testJSON(t *testing.T, j string, exp any) {
	t.Run(j, func(t *testing.T) {
		x, err := ParseString(JSON{}, j)
		require.NoError(t, err)

		assert.Equal(t, exp, x)
	})
}
