package core

import (
	"os"
	"runtime"
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
}

// New returns a configured Logger with the given options applied.
// Defaults: level=Trace (all entries pass), errors silently discarded.
// If no formatter is set but outputs are present, the text formatter is used.
func New(opts ...Option) *Logger {
	cfg := Config{
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
		f, err := fileout.New(path)
		if err != nil {
			cfg.ErrorHandler(err)
			continue
		}
		cfg.Outputs = append(cfg.Outputs, f)
	}
	// Auto-select formatter when outputs exist but none was explicitly configured.
	// Priority for pretty: forcePretty flag > TTY console detection.
	// Priority for theme:  WithTheme > console.WithStyles > styles.Default().
	if cfg.Formatter == nil && len(cfg.Outputs) > 0 {
		var ttyConsole *console.Console
		for _, o := range cfg.Outputs {
			if con, ok := o.(*console.Console); ok && con.IsTTY {
				ttyConsole = con
				break
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
func (l *Logger) SetLevel(lvl level.Level) {
	l.atomicLevel.Store(uint32(lvl))
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

// Fatal logs at FATAL level, then calls os.Exit(1).
func (l *Logger) Fatal(msg string, fields ...interfaces.Field) {
	l.log(level.Fatal, msg, fields...)
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
func (l *Logger) Close() {
	for _, out := range l.cfg.Outputs {
		if err := out.Close(); err != nil {
			l.cfg.ErrorHandler(err)
		}
	}
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
		// Skip frames: runtime.Callers(0), capturedCaller(1), log(2), <level method>(3)
		// → frame 4 is the user call site.
		entry.Caller = capturedCaller(4)
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

// capturedCaller returns call-site info by skipping skip frames on the stack.
func capturedCaller(skip int) interfaces.Caller {
	var pcs [1]uintptr
	if runtime.Callers(skip, pcs[:]) < 1 {
		return interfaces.Caller{}
	}
	frame, _ := runtime.CallersFrames(pcs[:]).Next()
	return interfaces.Caller{
		File:     frame.File,
		Line:     frame.Line,
		Function: frame.Function,
	}
}
