package buffer_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tidjee-dev/tlog/internal/buffer"
)

func TestGet_ReturnsEmptyBuffer(t *testing.T) {
	buf := buffer.Get()
	require.NotNil(t, buf)
	assert.Equal(t, 0, buf.Len())
	buffer.Put(buf)
}

func TestGet_ResetOnReuse(t *testing.T) {
	buf := buffer.Get()
	buf.WriteString("hello")
	assert.Equal(t, 5, buf.Len())
	buffer.Put(buf)

	buf2 := buffer.Get()
	// May or may not be the same pointer, but must be empty.
	assert.Equal(t, 0, buf2.Len())
	buffer.Put(buf2)
}

func TestGetPut_WriteAndRead(t *testing.T) {
	buf := buffer.Get()
	buf.WriteString("tlog buffer")
	assert.Equal(t, "tlog buffer", buf.String())
	buffer.Put(buf)
}

func TestGetPut_Concurrent(t *testing.T) {
	const goroutines = 50
	const writes = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			for range writes {
				buf := buffer.Get()
				buf.WriteString("concurrent write")
				_ = buf.Len()
				buffer.Put(buf)
			}
		}()
	}
	wg.Wait()
}
