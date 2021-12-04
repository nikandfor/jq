package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvalMath(t *testing.T) {
	testEvalMath(t, `1 + 1`, int64(2), nil)

	testEvalMath(t, `2 * 3`, int64(6), nil)

	testEvalMath(t, `2 * x`, int64(6), int64(3))

	testEvalMath(t, `2 * x`, int64(6), map[string]interface{}{
		"x": int64(3),
	})
}

func testEvalMath(t *testing.T, q string, exp, ctx any) {
	testEval(t, Math{}, q, exp, ctx)
}

func testEval(t *testing.T, p Parser, q string, exp, ctx any) {
	t.Run(q, func(t *testing.T) {
		x, err := Parse(p, []byte(q))
		require.NoError(t, err)

		y, err := Build(ctx, x)
		assert.NoError(t, err)

		z, err := Eval(ctx, y)
		assert.NoError(t, err)

		assert.Equal(t, exp, z)
	})
}
