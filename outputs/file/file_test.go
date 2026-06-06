package file

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPath(base, appName, file string) string {
	return filepath.Join(base, appName, file)
}

func TestWriteAndClose(t *testing.T) {
	base := t.TempDir()
	appName := "tlog-demo"
	path := "test-write.log"

	f, err := NewWithBase(base, appName, path)
	require.NoError(t, err)

	require.NoError(t, f.Write([]byte("line1\n")))
	require.NoError(t, f.Write([]byte("line2\n")))
	require.NoError(t, f.Close())

	got, err := os.ReadFile(testPath(base, appName, path))
	require.NoError(t, err)

	assert.Equal(t, "line1\nline2\n", string(got))
}

func TestCloseFlushesBuffer(t *testing.T) {
	base := t.TempDir()
	appName := "tlog-demo"
	path := "test-flush.log"

	f, err := NewWithBase(base, appName, path)
	require.NoError(t, err)

	require.NoError(t, f.Write([]byte("buffered\n")))
	require.NoError(t, f.Close())

	got, err := os.ReadFile(testPath(base, appName, path))
	require.NoError(t, err)

	assert.Equal(t, "buffered\n", string(got))
}

func TestAppendMode(t *testing.T) {
	base := t.TempDir()
	appName := "tlog-demo"
	path := "test-append.log"

	f, err := NewWithBase(base, appName, path)
	require.NoError(t, err)

	require.NoError(t, f.Write([]byte("first\n")))
	require.NoError(t, f.Close())

	f, err = NewWithBase(base, appName, path)
	require.NoError(t, err)

	require.NoError(t, f.Write([]byte("second\n")))
	require.NoError(t, f.Close())

	got, err := os.ReadFile(testPath(base, appName, path))
	require.NoError(t, err)

	assert.Equal(t, "first\nsecond\n", string(got))
}

func TestNewErrorInvalidPath(t *testing.T) {
	_, err := New("tlog-demo", "/nonexistent/directory/app.log")

	assert.Error(t, err)
}

func TestConcurrentWrites(t *testing.T) {
	base := t.TempDir()
	appName := "tlog-demo"
	path := "test-concurrent.log"

	f, err := NewWithBase(base, appName, path)
	require.NoError(t, err)

	const goroutines = 50
	const msg = "log line\n"

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			require.NoError(t, f.Write([]byte(msg)))
		}()
	}

	wg.Wait()

	require.NoError(t, f.Close())

	got, err := os.ReadFile(testPath(base, appName, path))
	require.NoError(t, err)

	assert.Equal(t, goroutines*len(msg), len(got))
}
