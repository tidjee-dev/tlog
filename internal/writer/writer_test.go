package writer_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidjee-dev/tlog/internal/writer"
)

func TestWriter_WriteAndFlush(t *testing.T) {
	var buf bytes.Buffer
	w := writer.New(&buf)

	n, err := w.Write([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)

	// Data may still be in the buffer; flush to push it through.
	require.NoError(t, w.Flush())
	assert.Equal(t, "hello", buf.String())
}

func TestWriter_ConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	w := writer.New(&buf)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = w.Write([]byte("x"))
		}()
	}
	wg.Wait()
	require.NoError(t, w.Flush())
	assert.Equal(t, 50, len(buf.String()))
}

func TestWriter_NewSize(t *testing.T) {
	var buf bytes.Buffer
	w := writer.NewSize(&buf, 1024)

	_, err := w.Write([]byte("sized"))
	require.NoError(t, err)
	require.NoError(t, w.Flush())
	assert.Equal(t, "sized", buf.String())
}
