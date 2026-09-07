package console

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidjee-dev/tlog/styles"
)

func TestWriteToBuffer(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithWriter(&buf))

	err := c.Write([]byte("hello\n"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", buf.String())
}

func TestWriteMultiple(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithWriter(&buf))

	require.NoError(t, c.Write([]byte("line1\n")))
	require.NoError(t, c.Write([]byte("line2\n")))
	assert.Equal(t, "line1\nline2\n", buf.String())
}

func TestCloseIsNoop(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithWriter(&buf))
	assert.NoError(t, c.Close())
}

func TestIsTTYFalseForBuffer(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithWriter(&buf))
	assert.False(t, c.IsTTY)
}

func TestWithTTYOverride(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithWriter(&buf), WithTTY(true))
	assert.True(t, c.IsTTY)
}

func TestWithTTYFalseOverride(t *testing.T) {
	// Even if something were a real TTY, WithTTY(false) must override it.
	c := New(WithTTY(false))
	assert.False(t, c.IsTTY)
}

func TestWithStylesStored(t *testing.T) {
	c := New(WithStyles(styles.Dev()))
	s, ok := c.Styles()
	assert.True(t, ok)
	_ = s // styles.Styles is a value type; presence is the assertion
}

func TestWithStylesNotSetByDefault(t *testing.T) {
	c := New()
	_, ok := c.Styles()
	assert.False(t, ok)
}

func TestWithStylesMinimal(t *testing.T) {
	c := New(WithStyles(styles.Minimal()))
	_, ok := c.Styles()
	assert.True(t, ok)
}

func TestConcurrentWrites(t *testing.T) {
	var buf safeBuffer
	c := New(WithWriter(&buf))

	const goroutines = 50
	const msg = "log line\n"

	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			// require.* must only run in the test goroutine;
			// report via channel and assert after Wait.
			errs <- c.Write([]byte(msg))
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		assert.NoError(t, err)
	}

	assert.Equal(t, goroutines*len(msg), buf.Len())
}

func TestWriteNilWriterReturnsError(t *testing.T) {
	c := New(WithWriter(nil))
	assert.Error(t, c.Write([]byte("hello\n")), "nil writer must error, not panic")
}

// safeBuffer wraps bytes.Buffer with a mutex so the test itself is race-free
// when checking Len() after concurrent writes to the Console.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Len()
}
