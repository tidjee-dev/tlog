package tlog

import "sync/atomic"

// _defaultLogger holds the package-level default logger.
// Initialized to a no-op logger in init() so package-level calls are always safe.
var _defaultLogger atomic.Pointer[Logger]

func init() {
	noop := New()
	_defaultLogger.Store(noop)
}

// SetDefault sets the package-level default logger used by the top-level
// convenience functions (Trace, Debug, Info, Warn, Error).
// Safe to call concurrently. Replaces any previously set default.
// Panics if log is nil.
//
// Warning: global mutable state makes code harder to test and reason about.
// Prefer passing loggers explicitly via function parameters or context.
// Not recommended for use in library code.
func SetDefault(log *Logger) {
	if log == nil {
		panic("tlog: SetDefault called with nil logger")
	}
	_defaultLogger.Store(log)
}

// Default returns the current package-level default logger.
// The initial default is a no-op logger (no outputs, no formatter).
func Default() *Logger {
	return _defaultLogger.Load()
}

// Trace logs at TRACE level using the default logger.
func Trace(msg string, fields ...Field) { Default().Trace(msg, fields...) }

// Debug logs at DEBUG level using the default logger.
func Debug(msg string, fields ...Field) { Default().Debug(msg, fields...) }

// Info logs at INFO level using the default logger.
func Info(msg string, fields ...Field) { Default().Info(msg, fields...) }

// Warn logs at WARN level using the default logger.
func Warn(msg string, fields ...Field) { Default().Warn(msg, fields...) }

// Error logs at ERROR level using the default logger.
func Error(msg string, fields ...Field) { Default().Error(msg, fields...) }
