package core

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/tidjee-dev/tlog/formatter/pretty"
	"github.com/tidjee-dev/tlog/formatter/text"
	"github.com/tidjee-dev/tlog/interfaces"
	"github.com/tidjee-dev/tlog/internal/clock"
	"github.com/tidjee-dev/tlog/level"
	"github.com/tidjee-dev/tlog/outputs/console"
	fileout "github.com/tidjee-dev/tlog/outputs/file"
	"github.com/tidjee-dev/tlog/styles"
)

// Logger dispatches log entries to one or more outputs via a formatter.
// All fields are immutable after construction — Logger is safe for concurrent use.
// The minimum level can be changed at any time via SetLevel; reads are atomic.
type Logger struct {
	cfg         Config
	atomicLevel atomic.Uint32
	closeOnce   sync.Once
}

// Compile-time checks: built-in clocks satisfy the public interfaces.Clock,
// so existing WithClock(clock.Real{}) / WithClock(clock.NewMock(t)) callers
// keep compiling after the parameter was widened from clock.Clock.
var (
	_ interfaces.Clock = clock.Real{}
	_ interfaces.Clock = (*clock.Mock)(nil)
)

// New returns a configured Logger with the given options applied.
// Defaults: level=Trace (all entries pass), errors silently discarded.
// If no formatter is set but outputs are present, the text formatter is used.
func New(opts ...Option) *Logger {
	cfg := Config{
		AppName:      "tlog-dev-demo", // default app name for file outputs; can be overridden by WithFile or custom file output
		Level:        level.Trace,
		Clock:        clock.Real{},
		ErrorHandler: func(error) {},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	// Open any file paths registered via WithFile. The error handler is already
	// set at this point (all options have been applied), so failures are routed
	// correctly regardless of option order.
	for _, path := range cfg.filePaths {
		f, err := fileout.New(cfg.AppName, path)
		if err != nil {
			cfg.ErrorHandler(err)
			continue
		}
		cfg.Outputs = append(cfg.Outputs, f)
	}
	// Auto-select formatter when outputs exist but none was explicitly configured.
	// Priority for pretty: forcePretty flag > TTY console detection, unless
	// colour is suppressed by the environment (non-empty NO_COLOR or
	// TERM=dumb), which downgrades TTY auto-detection to plain text.
	// An explicit WithPretty still forces styled output.
	// Priority for theme:  WithTheme > console.WithStyles > styles.Default().
	if cfg.Formatter == nil && len(cfg.Outputs) > 0 {
		var ttyConsole *console.Console
		if !colorSuppressed() {
			for _, o := range cfg.Outputs {
				if con, ok := o.(*console.Console); ok && con.TTY() {
					ttyConsole = con
					break
				}
			}
		}

		usePretty := cfg.forcePretty || ttyConsole != nil

		if usePretty {
			var prettyOpts []pretty.Option
			if cfg.TimestampFormat != "" {
				prettyOpts = append(prettyOpts, pretty.WithTimestampFormat(cfg.TimestampFormat))
			}
			// Resolve theme: WithTheme > console.WithStyles > nil (pretty uses Default)
			var theme *styles.Styles
			if cfg.prettyTheme != nil {
				theme = cfg.prettyTheme
			} else if ttyConsole != nil {
				if s, ok := ttyConsole.Styles(); ok {
					theme = &s
				}
			}
			// Apply WithStyles modifier on top of resolved theme.
			if cfg.stylesModifier != nil {
				if theme == nil {
					s := styles.Dev()
					theme = &s
				}
				cfg.stylesModifier(theme)
			}
			if theme != nil {
				prettyOpts = append(prettyOpts, pretty.WithStyles(*theme))
			}
			if cfg.forcePretty {
				prettyOpts = append(prettyOpts, pretty.WithTTY(true))
			}
			cfg.Formatter = pretty.New(prettyOpts...)
		} else {
			if cfg.TimestampFormat != "" {
				cfg.Formatter = text.New(text.WithTimestampFormat(cfg.TimestampFormat))
			} else {
				cfg.Formatter = text.New()
			}
		}
	}
	l := &Logger{cfg: cfg}
	l.atomicLevel.Store(uint32(cfg.Level))
	return l
}

// SetLevel changes the minimum log level for this Logger.
// The change is visible to all concurrent callers immediately.
// It does not propagate to existing With/WithLevel children (they snapshotted
// the level at creation); read the live value with Level().
func (l *Logger) SetLevel(lvl level.Level) {
	l.atomicLevel.Store(uint32(lvl))
}

// Level returns the live minimum log level (the atomic value, not the
// initial Config). Entries below it are dropped.
func (l *Logger) Level() level.Level {
	return level.Level(l.atomicLevel.Load())
}

// Enabled reports whether an entry at lvl would be logged.
// Use it to skip expensive field construction.
func (l *Logger) Enabled(lvl level.Level) bool {
	return lvl >= l.Level()
}

// With returns a new child Logger that inherits all settings from the parent
// and prepends the parent's fields before any fields supplied at call sites.
// The parent Logger is not modified.
func (l *Logger) With(fields ...interfaces.Field) *Logger {
	merged := make([]interfaces.Field, 0, len(l.cfg.Fields)+len(fields))
	merged = append(merged, l.cfg.Fields...)
	merged = append(merged, fields...)
	child := l.cfg
	child.Fields = merged
	childLogger := &Logger{cfg: child}
	childLogger.atomicLevel.Store(l.atomicLevel.Load())
	return childLogger
}

// WithLevel returns a new child Logger with the minimum log level overridden.
// All other settings (outputs, formatter, fields) are inherited from the parent.
// The parent Logger is not modified.
func (l *Logger) WithLevel(lvl level.Level) *Logger {
	child := l.cfg
	child.Level = lvl
	childLogger := &Logger{cfg: child}
	childLogger.atomicLevel.Store(uint32(lvl))
	return childLogger
}

// Trace logs at TRACE level.
func (l *Logger) Trace(msg string, fields ...interfaces.Field) {
	l.log(level.Trace, msg, fields...)
}

// Debug logs at DEBUG level.
func (l *Logger) Debug(msg string, fields ...interfaces.Field) {
	l.log(level.Debug, msg, fields...)
}

// Info logs at INFO level.
func (l *Logger) Info(msg string, fields ...interfaces.Field) {
	l.log(level.Info, msg, fields...)
}

// Warn logs at WARN level.
func (l *Logger) Warn(msg string, fields ...interfaces.Field) {
	l.log(level.Warn, msg, fields...)
}

// Error logs at ERROR level.
func (l *Logger) Error(msg string, fields ...interfaces.Field) {
	l.log(level.Error, msg, fields...)
}

// syncer is an optional Output hook for durability.
// Outputs that buffer data should implement Sync() error;
// Fatal calls it best-effort before os.Exit.
type syncer interface {
	Sync() error
}

// Fatal logs at FATAL level, best-effort syncs outputs that implement
// Sync() error, then calls os.Exit(1). Deferred funcs (including Close)
// do not run after os.Exit, so the Sync step is what preserves the last
// line for buffered outputs. Sync errors are routed to the error handler.
func (l *Logger) Fatal(msg string, fields ...interfaces.Field) {
	l.log(level.Fatal, msg, fields...)
	for _, out := range l.cfg.Outputs {
		if s, ok := out.(syncer); ok {
			if err := s.Sync(); err != nil {
				l.cfg.ErrorHandler(err)
			}
		}
	}
	os.Exit(1)
}

// Panic logs at PANIC level, then panics with msg.
func (l *Logger) Panic(msg string, fields ...interfaces.Field) {
	l.log(level.Panic, msg, fields...)
	panic(msg)
}

// Close closes all outputs associated with this logger.
// It should be called when the logger is no longer needed to flush and release
// any underlying resources (e.g. open files).
// Errors from individual outputs are routed to the error handler.
// Close is idempotent per Logger (repeat calls are no-ops). With/WithLevel
// children share the parent's outputs, so close only once — typically the
// parent or the explicitly owned child — to avoid double-Close on custom
// outputs. Built-in file output is idempotent and console Close is a no-op.
func (l *Logger) Close() {
	l.closeOnce.Do(func() {
		for _, out := range l.cfg.Outputs {
			if err := out.Close(); err != nil {
				l.cfg.ErrorHandler(err)
			}
		}
	})
}

// log is the internal dispatch path shared by all level methods.
func (l *Logger) log(lvl level.Level, msg string, fields ...interfaces.Field) {
	atomicLevel := l.atomicLevel.Load()
	if lvl < level.Level(atomicLevel) {
		return
	}
	if l.cfg.Formatter == nil {
		return
	}

	// Merge parent fields with call-site fields (parent first).
	allFields := fields
	if len(l.cfg.Fields) > 0 {
		allFields = make([]interfaces.Field, 0, len(l.cfg.Fields)+len(fields))
		allFields = append(allFields, l.cfg.Fields...)
		allFields = append(allFields, fields...)
	}

	entry := interfaces.Entry{
		Timestamp: l.cfg.Clock.Now(),
		Level:     lvl,
		Message:   msg,
		Fields:    allFields,
	}

	if l.cfg.CallerEnabled {
		entry.Caller = capturedCaller(l.cfg.CallerSkip)
	}

	b, err := l.cfg.Formatter.Format(entry)
	if err != nil {
		l.cfg.ErrorHandler(err)
		return
	}

	for _, out := range l.cfg.Outputs {
		if err := out.Write(b); err != nil {
			l.cfg.ErrorHandler(err)
		}
	}
}

// colorSuppressed reports whether the environment asks for no colour:
// a non-empty NO_COLOR (https://no-color.org) or TERM=dumb.
func colorSuppressed() bool {
	if v, ok := os.LookupEnv("NO_COLOR"); ok && v != "" {
		return true
	}
	return os.Getenv("TERM") == "dumb"
}

// capturedCaller returns the first non-tlog call-site frame, skipping
// extraSkip additional user frames (e.g. helper wrappers configured via
// WithCallerSkip). Only the known logger-internal frames are skipped
// (log dispatch, level methods, package helpers), so package-level helpers
// and white-box tests inside the module still attribute correctly.
func capturedCaller(extraSkip int) interfaces.Caller {
	const maxFrames = 16
	var pcs [maxFrames]uintptr
	// Skip runtime.Callers + capturedCaller itself; the rest is filtered below.
	n := runtime.Callers(2, pcs[:])
	if n == 0 {
		return interfaces.Caller{}
	}
	frames := runtime.CallersFrames(pcs[:n])
	skipped := 0
	for {
		frame, more := frames.Next()
		if isInternalCaller(frame.Function) {
			if !more {
				return interfaces.Caller{}
			}
			continue
		}
		if skipped < extraSkip {
			skipped++
			if !more {
				return interfaces.Caller{}
			}
			continue
		}
		return interfaces.Caller{
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
		}
	}
}

// isInternalCaller reports whether fn is a tlog log-path frame that must be
// skipped during caller attribution. It matches exact internal symbols only,
// not the whole module, so user code and tests living in the module (e.g.
// package core tests) are still reported as call sites.
func isInternalCaller(fn string) bool {
	if fn == "github.com/tidjee-dev/tlog/core.capturedCaller" ||
		fn == "github.com/tidjee-dev/tlog/core.isInternalCaller" ||
		fn == "github.com/tidjee-dev/tlog/core.(*Logger).log" {
		return true
	}
	if strings.HasPrefix(fn, "github.com/tidjee-dev/tlog/core.(*Logger).") {
		// Trace/Debug/Info/Warn/Error/Fatal/Panic + Close/With helpers.
		return true
	}
	switch fn {
	case "github.com/tidjee-dev/tlog.Trace",
		"github.com/tidjee-dev/tlog.Debug",
		"github.com/tidjee-dev/tlog.Info",
		"github.com/tidjee-dev/tlog.Warn",
		"github.com/tidjee-dev/tlog.Error",
		"github.com/tidjee-dev/tlog.Default":
		return true
	}
	return false
}
