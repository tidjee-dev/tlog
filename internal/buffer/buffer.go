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

// maxPooledCap bounds pool retention: buffers grown by a single huge log
// line (e.g. 1 MiB Any payload) are dropped instead of pinned in sync.Pool.
const maxPooledCap = 64 * 1024

// Get returns a reset buffer from the pool.
func Get() *bytes.Buffer {
	b := pool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

// Put returns a buffer to the pool.
// The buffer must not be used after this call.
// Buffers over maxPooledCap are dropped to avoid pinning huge allocations.
func Put(b *bytes.Buffer) {
	if b == nil {
		return
	}
	if b.Cap() > maxPooledCap {
		return
	}
	pool.Put(b)
}
