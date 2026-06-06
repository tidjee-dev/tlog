package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// File is an Output that appends log lines to a file.
// Writes are buffered via internal/writer and the buffer is flushed on Close.
// It is thread-safe.
type File struct {
	mu sync.Mutex
	f  *os.File
	// bw *writer.Writer
}

// New opens or creates a log file for appending inside the default user logs directory.
//
// The default base directory is:
//
//	~/.logs/<appName>/<path>
//
// Examples:
//
//	~/.logs/myapp/app.log
//	~/.logs/myapp/api/dev.log
//
// Permissions:
//   - directories: 0700
//   - files: 0600

func New(appName string, path string) (*File, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return NewWithBase(filepath.Join(home, ".logs"), appName, path)
}

// NewWithBase opens or creates a log file for appending inside a given base directory.
//
// The final resolved path is:
//
//	baseDir/<appName>/<path>
//
// Examples:
//
//	/base/.logs/myapp/app.log
//	/base/.logs/myapp/api/dev.log
//
// Security:
//   - absolute paths are rejected
//   - path traversal outside baseDir is prevented
//
// Permissions:
//   - directories: 0700
//   - files: 0600
func NewWithBase(baseDir, appName, path string) (*File, error) {
	if filepath.IsAbs(path) {
		return nil, errors.New("absolute paths are not allowed")
	}

	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(baseDir, appName, cleanPath)

	rel, err := filepath.Rel(filepath.Join(baseDir, appName), fullPath)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(rel, "..") {
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

// Write implements interfaces.Output.
// Thread-safety is provided by the underlying writer.Writer.
func (f *File) Write(p []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, err := f.f.Write(p)
	return err
}

// Close implements interfaces.Output.
// It closes the underlying file.
func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.f.Close()
}
