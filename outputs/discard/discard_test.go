package discard

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteReturnsNil(t *testing.T) {
	d := New()
	require.NoError(t, d.Write([]byte("anything")))
}

func TestWriteDropsData(t *testing.T) {
	d := New()
	// Multiple writes must all succeed silently.
	for range 100 {
		require.NoError(t, d.Write([]byte("log line\n")))
	}
}

func TestCloseIsNoop(t *testing.T) {
	d := New()
	assert.NoError(t, d.Close())
}
