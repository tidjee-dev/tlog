package console

import (
	"io"
	"os"
	"sync"

	"github.com/mattn/go-isatty"
	"github.com/tidjee-dev/tlog/styles"
)

// Console is an Output that writes to a writer (default: os.Stdout).
// It is thread-safe. TTY detection is performed once at construction.
// IsTTY must not be mutated after the Console is shared; configure it via
// New(WithTTY(...)) / New(WithWriter(...)) and read it via TTY().
type Console struct {
	mu        sync.RWMutex
	w         io.Writer
	IsTTY     bool
	theme     styles.Styles
	hasStyles bool
}

// Option configures a Console.
type Option func(*Console)

// WithWriter replaces the default os.Stdout with the given writer.
// TTY detection is still attempted on the writer if it is an *os.File.
func WithWriter(w io.Writer) Option {
	return func(c *Console) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.w = w
		c.IsTTY = isTTY(w)
	}
}

// WithStyles sets a Styles override used by the pretty formatter when the
// logger auto-selects it for TTY output. Has no effect if a formatter is
// configured explicitly via WithFormatter.
func WithStyles(s styles.Styles) Option {
	return func(c *Console) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.theme = s
		c.hasStyles = true
	}
}

// WithTTY overrides TTY detection. Useful in tests or when the caller knows
// the terminal state without relying on isatty.
func WithTTY(tty bool) Option {
	return func(c *Console) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.IsTTY = tty
	}
}

// TTY reports the cached TTY state in a thread-safe way.
// Prefer it over reading IsTTY directly when the Console may be shared.
func (c *Console) TTY() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IsTTY
}

// Styles returns the Styles override and true if one was set via WithStyles,
// or an empty Styles and false otherwise.
func (c *Console) Styles() (styles.Styles, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.theme, c.hasStyles
}

// New returns a Console that writes to os.Stdout by default.
// TTY detection is evaluated once here and cached.
func New(opts ...Option) *Console {
	c := &Console{
		w:     os.Stdout,
		IsTTY: isTTY(os.Stdout),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Write implements interfaces.Output.
// The call is serialised with a mutex so concurrent loggers are safe.
// It loops until all bytes are written; a short write without an error
// is reported as io.ErrShortWrite.
func (c *Console) Write(p []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.w == nil {
		return io.ErrClosedPipe
	}
	total := 0
	for total < len(p) {
		n, err := c.w.Write(p[total:])
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

// Close implements interfaces.Output.
// Stdout is not closeable; this is a no-op and always returns nil.
func (c *Console) Close() error {
	return nil
}

// Sync implements an optional flush hook.
// Console writes are unbuffered, so this is a no-op and always returns nil.
// It lets Fatal best-effort flush outputs that implement Sync() error.
func (c *Console) Sync() error { return nil }

// isTTY reports whether w is a TTY file descriptor.
func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
