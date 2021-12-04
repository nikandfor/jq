package jq

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJQ(t *testing.T) {
	testJQ(t, ".", `"a"`, `"a"`)
}

func testJQ(t *testing.T, q string, in, exp string) {
	var b bytes.Buffer

	w, err := Compile(&b, q)
	require.NoError(t, err)

	n, err := w.Write([]byte(in))
	require.NoError(t, err)
	assert.Equal(t, len(exp), n)
}
