// Package writer provides a thread-safe buffered writer with explicit flush.
// It wraps bufio.Writer behind a mutex so concurrent callers are safe,
// and exposes Flush to push buffered bytes to the underlying io.Writer.
package writer

import (
	"bufio"
	"io"
	"sync"
)

// Writer is a thread-safe buffered writer.
type Writer struct {
	mu sync.Mutex
	bw *bufio.Writer
}

// New wraps w in a Writer with the default buffer size.
func New(w io.Writer) *Writer {
	return &Writer{bw: bufio.NewWriter(w)}
}

// NewSize wraps w in a Writer with a buffer of size bytes.
func NewSize(w io.Writer, size int) *Writer {
	return &Writer{bw: bufio.NewWriterSize(w, size)}
}

// Write writes p to the buffer. Thread-safe.
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	n, err := w.bw.Write(p)
	w.mu.Unlock()
	return n, err
}

// Flush flushes buffered bytes to the underlying writer. Thread-safe.
func (w *Writer) Flush() error {
	w.mu.Lock()
	err := w.bw.Flush()
	w.mu.Unlock()
	return err
}
