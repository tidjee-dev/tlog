package file

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAndClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	f, err := New(path)
	require.NoError(t, err)

	require.NoError(t, f.Write([]byte("line1\n")))
	require.NoError(t, f.Write([]byte("line2\n")))
	require.NoError(t, f.Close())

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "line1\nline2\n", string(got))
}

func TestCloseFlushesBuffer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flush.log")
	f, err := New(path)
	require.NoError(t, err)

	require.NoError(t, f.Write([]byte("buffered\n")))

	// Before Close, bufio may not have flushed to disk.
	// After Close, content must be present.
	require.NoError(t, f.Close())

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "buffered\n", string(got))
}

func TestAppendMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "append.log")

	// First session.
	f, err := New(path)
	require.NoError(t, err)
	require.NoError(t, f.Write([]byte("first\n")))
	require.NoError(t, f.Close())

	// Second session — must append, not truncate.
	f, err = New(path)
	require.NoError(t, err)
	require.NoError(t, f.Write([]byte("second\n")))
	require.NoError(t, f.Close())

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "first\nsecond\n", string(got))
}

func TestNewErrorInvalidPath(t *testing.T) {
	_, err := New("/nonexistent/directory/app.log")
	assert.Error(t, err)
}

func TestConcurrentWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "concurrent.log")
	f, err := New(path)
	require.NoError(t, err)

	const goroutines = 50
	const msg = "log line\n"

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			require.NoError(t, f.Write([]byte(msg)))
		}()
	}
	wg.Wait()
	require.NoError(t, f.Close())

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, goroutines*len(msg), len(got))
}
