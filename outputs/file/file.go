package file

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// File is an Output that appends log lines to a file.
type File struct {
	mu     sync.Mutex
	f      *os.File
	closed bool
}

// New opens or creates a log file inside ~/.logs/<appName>/<path>.
func New(appName string, path string) (*File, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	base := filepath.Join(home, ".logs")
	return NewWithBase(base, appName, path)
}

// NewWithBase opens or creates a log file inside baseDir/<appName>/<path>.
// It prevents absolute paths and directory traversal.
func NewWithBase(baseDir, appName, path string) (*File, error) {
	if path == "" {
		return nil, errors.New("empty path is not allowed")
	}

	if filepath.IsAbs(path) {
		return nil, errors.New("absolute paths are not allowed")
	}

	cleanPath := filepath.Clean(path)

	baseAppDir := filepath.Join(baseDir, appName)
	fullPath := filepath.Join(baseAppDir, cleanPath)

	// Stronger traversal protection
	rel, err := filepath.Rel(baseAppDir, fullPath)
	if err != nil {
		return nil, err
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, errors.New("path escapes log directory")
	}

	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	f, err := os.OpenFile( // #nosec G304 -- path is validated and constrained to baseDir
		fullPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o600,
	)
	if err != nil {
		return nil, err
	}

	return &File{f: f}, nil
}

// Write appends data to the file in a thread-safe way.
// It loops until all bytes are written; a short write without an error
// is reported as io.ErrShortWrite so callers can route it to the error handler.
func (f *File) Write(p []byte) error {
	if f == nil {
		return errors.New("nil file")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return errors.New("write on closed file")
	}

	total := 0
	for total < len(p) {
		n, err := f.f.Write(p[total:])
		total += n
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}

// Close closes the file safely (idempotent).
func (f *File) Close() error {
	if f == nil {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true

	syncErr := f.f.Sync()
	closeErr := f.f.Close()
	return errors.Join(syncErr, closeErr)
}
