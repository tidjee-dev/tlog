// Package tlog is a lightweight, extensible, production-oriented logger for Go.
// It provides a minimal API with structured typed fields, human-readable console
// output, and JSON output compatible with log aggregators.
//
// Quick start:
//
//	log := tlog.New(tlog.WithConsole())
//	log.Info("server started", tlog.String("host", "localhost"), tlog.Int("port", 8080))
package tlog

import (
	"context"
	"time"

	"github.com/tidjee-dev/tlog/core"
	jsonfmt "github.com/tidjee-dev/tlog/formatter/json"
	"github.com/tidjee-dev/tlog/interfaces"
	"github.com/tidjee-dev/tlog/level"
	"github.com/tidjee-dev/tlog/outputs/console"
	"github.com/tidjee-dev/tlog/styles"
)

// Logger is the tlog logger. Safe for concurrent use.
type Logger = core.Logger

// Option configures a Logger.
type Option = core.Option

// Field is a typed key-value pair attached to a log entry.
type Field = interfaces.Field

// New creates a new Logger with the given options applied.
func New(opts ...Option) *Logger {
	return core.New(opts...)
}

// --- Options ----------------------------------------------------------------

// WithLevel sets the minimum log level. Entries below this level are dropped.
func WithLevel(l level.Level) Option { return core.WithLevel(l) }

// WithConsole adds a console (os.Stdout) output. When the terminal is a TTY
// the pretty formatter is selected automatically; otherwise plain text is used.
// Optional console.Option values (e.g. console.WithStyles, console.WithWriter)
// can be passed to customise the console output.
func WithConsole(opts ...console.Option) Option { return core.WithConsole(opts...) }

// WithOutput appends an output to the logger.
func WithOutput(o interfaces.Output) Option { return core.WithOutput(o) }

// WithFormatter sets the formatter used to render log entries.
func WithFormatter(f interfaces.Formatter) Option { return core.WithFormatter(f) }

// WithErrorHandler sets a function called when a formatter or output error occurs.
// The handler runs inline on the caller's goroutine: it must be non-blocking,
// goroutine-safe, and must not call back into the logger. A nil fn is ignored.
func WithErrorHandler(fn func(error)) Option { return core.WithErrorHandler(fn) }

// WithCaller enables caller info (file, line, function) in log entries.
func WithCaller() Option { return core.WithCaller() }

// WithCallerSkip enables caller info and skips extra frames beyond the
// logger internals (e.g. your own helper wrappers).
func WithCallerSkip(skip int) Option { return core.WithCallerSkip(skip) }

// WithTimestampFormat sets the time layout used by the default text formatter.
func WithTimestampFormat(format string) Option { return core.WithTimestampFormat(format) }

// WithJSON sets the JSON formatter.
// Optional jsonfmt.Option values (e.g. jsonfmt.WithTimeFormat) can be passed.
func WithJSON(opts ...jsonfmt.Option) Option { return core.WithJSON(opts...) }

// WithFile adds a file output that appends to the given path.
// The file is created if it does not exist. If opening fails, the error is
// routed to the error handler.
func WithFile(appName string, path string) Option { return core.WithFile(appName, path) }

// Styles is the set of lipgloss styles used by the pretty formatter.
// Use it as the parameter type when calling WithTheme or WithStyles.
type Styles = styles.Styles

// WithDiscard adds an output that silently drops all log entries.
// Useful in tests and benchmarks.
func WithDiscard() Option { return core.WithDiscard() }

// WithClock sets the time source used for entry timestamps.
// Useful in tests to inject a fixed or controllable clock.
// Accepts any interfaces.Clock; clock.Real and *clock.Mock both qualify.
func WithClock(clk interfaces.Clock) Option { return core.WithClock(clk) }

// WithTheme sets the complete Styles theme used by the auto-selected pretty
// formatter. Overrides any styles set via console.WithStyles.
func WithTheme(theme styles.Styles) Option { return core.WithTheme(theme) }

// WithStyles applies a modifier function to the active theme before the pretty
// formatter is created. Receives a pointer to a copy of the current theme so
// individual fields can be changed without replacing the whole theme.
func WithStyles(fn func(*styles.Styles)) Option { return core.WithStyles(fn) }

// WithPretty forces the lipgloss-styled pretty formatter regardless of whether
// the output is a TTY. Useful when piping output to a pager or tool that
// understands ANSI codes.
func WithPretty() Option { return core.WithPretty() }

// --- Context ----------------------------------------------------------------

// NewContext returns a new context carrying log.
// Retrieve it later with FromContext.
func NewContext(ctx context.Context, log *Logger) context.Context {
	return core.NewContext(ctx, log)
}

// FromContext returns the Logger stored in ctx by NewContext.
// If no logger is present, a no-op logger is returned (never nil).
func FromContext(ctx context.Context) *Logger {
	return core.FromContext(ctx)
}

// WithContext returns a copy of ctx whose embedded logger has fields appended.
// Shorthand for NewContext(ctx, FromContext(ctx).With(fields...)).
func WithContext(ctx context.Context, fields ...Field) context.Context {
	return core.WithContext(ctx, fields...)
}

// --- Field constructors -----------------------------------------------------

// String returns a Field with a string value.
func String(key, val string) Field { return interfaces.String(key, val) }

// Int returns a Field with an int value.
func Int(key string, val int) Field { return interfaces.Int(key, val) }

// Int64 returns a Field with an int64 value.
func Int64(key string, val int64) Field { return interfaces.Int64(key, val) }

// Float64 returns a Field with a float64 value.
func Float64(key string, val float64) Field { return interfaces.Float64(key, val) }

// Bool returns a Field with a bool value.
func Bool(key string, val bool) Field { return interfaces.Bool(key, val) }

// Duration returns a Field with a time.Duration value.
func Duration(key string, val time.Duration) Field { return interfaces.Duration(key, val) }

// Time returns a Field with a time.Time value.
func Time(key string, val time.Time) Field { return interfaces.Time(key, val) }

// Err returns a Field capturing an error under the key "error".
// If err is nil the value is stored as the string "<nil>".
func Err(err error) Field { return interfaces.Err(err) }

// Any returns a Field with an arbitrary value.
func Any(key string, val any) Field { return interfaces.Any(key, val) }
