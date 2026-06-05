// Package buffer provides a sync.Pool-backed bytes.Buffer pool to reduce
// allocations in hot logging paths.
package buffer

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// Get returns a reset buffer from the pool.
func Get() *bytes.Buffer {
	b := pool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

// Put returns a buffer to the pool.
// The buffer must not be used after this call.
func Put(b *bytes.Buffer) {
	pool.Put(b)
}
