package core

import (
	"github.com/tidjee-dev/tlog/formatter/json"
	"github.com/tidjee-dev/tlog/interfaces"
	"github.com/tidjee-dev/tlog/internal/clock"
	"github.com/tidjee-dev/tlog/level"
	"github.com/tidjee-dev/tlog/outputs/console"
	"github.com/tidjee-dev/tlog/outputs/discard"
	"github.com/tidjee-dev/tlog/styles"
)

// Config holds all configuration for a Logger.
// It is built once via functional options and never mutated after New returns.
type Config struct {
	Level           level.Level
	Clock           clock.Clock
	Outputs         []interfaces.Output
	Formatter       interfaces.Formatter
	ErrorHandler    func(error)
	CallerEnabled   bool
	TimestampFormat string
	Fields          []interfaces.Field
	filePaths       []string             // deferred file opens; processed by core.New
	prettyTheme     *styles.Styles       // set by WithTheme
	stylesModifier  func(*styles.Styles) // set by WithStyles
	forcePretty     bool                 // set by WithPretty
}

// Option is a functional option for configuring a Logger.
type Option func(*Config)

// WithLevel sets the minimum log level. Entries below this level are dropped.
func WithLevel(l level.Level) Option {
	return func(c *Config) {
		c.Level = l
	}
}

// WithClock sets the time source used for entry timestamps.
// The default is clock.Real{} which calls time.Now().
// Use clock.NewMock to control time deterministically in tests.
func WithClock(clk clock.Clock) Option {
	return func(c *Config) {
		c.Clock = clk
	}
}

// WithConsole adds a console output (os.Stdout) to the logger.
// If no formatter has been configured by the time New is called, a pretty
// formatter is used automatically for TTY output; otherwise text is used.
func WithConsole(opts ...console.Option) Option {
	return func(c *Config) {
		c.Outputs = append(c.Outputs, console.New(opts...))
	}
}

// WithOutput appends an output to the logger.
func WithOutput(o interfaces.Output) Option {
	return func(c *Config) {
		c.Outputs = append(c.Outputs, o)
	}
}

// WithFormatter sets the formatter used to render log entries.
func WithFormatter(f interfaces.Formatter) Option {
	return func(c *Config) {
		c.Formatter = f
	}
}

// WithErrorHandler sets a function called when a formatter or output error
// occurs. By default errors are silently discarded.
func WithErrorHandler(fn func(error)) Option {
	return func(c *Config) {
		c.ErrorHandler = fn
	}
}

// WithCaller enables caller info (file, line, function) in log entries.
func WithCaller() Option {
	return func(c *Config) {
		c.CallerEnabled = true
	}
}

// WithTimestampFormat sets the time layout used by the default text formatter.
// Has no effect if a custom formatter is set via WithFormatter.
func WithTimestampFormat(format string) Option {
	return func(c *Config) {
		c.TimestampFormat = format
	}
}

// WithJSON sets the JSON formatter. Any previously set formatter is replaced.
func WithJSON(opts ...json.Option) Option {
	return func(c *Config) {
		c.Formatter = json.New(opts...)
	}
}

// WithFile adds a file output that appends to the given path.
// The file is opened (or created) during New. If it cannot be opened, the
// error is routed to the error handler.
func WithFile(path string) Option {
	return func(c *Config) {
		c.filePaths = append(c.filePaths, path)
	}
}

// WithDiscard adds an output that silently drops all log entries.
// Useful in tests and benchmarks.
func WithDiscard() Option {
	return func(c *Config) {
		c.Outputs = append(c.Outputs, discard.New())
	}
}

// WithTheme sets the complete Styles theme used by the auto-selected pretty
// formatter. Overrides any styles set via console.WithStyles.
func WithTheme(theme styles.Styles) Option {
	return func(c *Config) {
		c.prettyTheme = &theme
	}
}

// WithStyles applies a modifier function to the active theme before the pretty
// formatter is created. The modifier receives a pointer to a copy of the
// current theme (WithTheme > console.WithStyles > Default) and may change any
// fields. Applied after WithTheme, so both can be combined.
func WithStyles(fn func(*styles.Styles)) Option {
	return func(c *Config) {
		c.stylesModifier = fn
	}
}

// WithPretty forces the lipgloss-styled pretty formatter regardless of whether
// the output is a TTY. Useful when piping output to a pager or another tool
// that understands ANSI.
func WithPretty() Option {
	return func(c *Config) {
		c.forcePretty = true
	}
}
