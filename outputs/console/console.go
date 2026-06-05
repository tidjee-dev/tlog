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
type Console struct {
	mu        sync.Mutex
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
		c.w = w
		c.IsTTY = isTTY(w)
	}
}

// WithStyles sets a Styles override used by the pretty formatter when the
// logger auto-selects it for TTY output. Has no effect if a formatter is
// configured explicitly via WithFormatter.
func WithStyles(s styles.Styles) Option {
	return func(c *Console) {
		c.theme = s
		c.hasStyles = true
	}
}

// WithTTY overrides TTY detection. Useful in tests or when the caller knows
// the terminal state without relying on isatty.
func WithTTY(tty bool) Option {
	return func(c *Console) {
		c.IsTTY = tty
	}
}

// Styles returns the Styles override and true if one was set via WithStyles,
// or an empty Styles and false otherwise.
func (c *Console) Styles() (styles.Styles, bool) {
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
func (c *Console) Write(p []byte) error {
	c.mu.Lock()
	_, err := c.w.Write(p)
	c.mu.Unlock()
	return err
}

// Close implements interfaces.Output.
// Stdout is not closeable; this is a no-op and always returns nil.
func (c *Console) Close() error {
	return nil
}

// isTTY reports whether w is a TTY file descriptor.
func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
