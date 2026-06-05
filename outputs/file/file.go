package file

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/tidjee-dev/tlog/internal/writer"
)

// File is an Output that appends log lines to a file.
// Writes are buffered via internal/writer and the buffer is flushed on Close.
// It is thread-safe.
type File struct {
	mu sync.Mutex
	f  *os.File
	bw *writer.Writer
}

// New opens (or creates) the file at path for appending and returns a File output.
// The file is created with mode 0600 if it does not exist.
// Returns an error if the file cannot be opened.
func New(path string) (*File, error) {
	baseDir := "./logs"

	clean := filepath.Clean(path)
	fullPath := filepath.Join(baseDir, clean)

	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	return &File{
		f:  f,
		bw: writer.New(f),
	}, nil
}

// Write implements interfaces.Output.
// Thread-safety is provided by the underlying writer.Writer.
func (f *File) Write(p []byte) error {
	_, err := f.bw.Write(p)
	return err
}

// Close implements interfaces.Output.
// It flushes the internal buffer and closes the underlying file.
func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.bw.Flush(); err != nil {
		_ = f.f.Close()
		return err
	}
	return f.f.Close()
}
