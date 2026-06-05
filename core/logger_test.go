package core

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidjee-dev/tlog/formatter/pretty"
	"github.com/tidjee-dev/tlog/formatter/text"
	"github.com/tidjee-dev/tlog/interfaces"
	"github.com/tidjee-dev/tlog/internal/clock"
	"github.com/tidjee-dev/tlog/level"
	"github.com/tidjee-dev/tlog/outputs/console"
	"github.com/tidjee-dev/tlog/styles"
)

// --------------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------------

// testOutput captures Write calls in a thread-safe buffer.
type testOutput struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (o *testOutput) Write(p []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	_, err := o.buf.Write(p)
	return err
}

func (o *testOutput) Close() error { return nil }

func (o *testOutput) String() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.buf.String()
}

func (o *testOutput) Len() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.buf.Len()
}

// errOutput always returns an error on Write.
type errOutput struct{ err error }

func (e *errOutput) Write([]byte) error { return e.err }
func (e *errOutput) Close() error       { return nil }

// newTestLogger returns a Logger writing to out using the text formatter.
func newTestLogger(out interfaces.Output, opts ...Option) *Logger {
	opts = append([]Option{WithOutput(out), WithFormatter(text.New())}, opts...)
	return New(opts...)
}

// --------------------------------------------------------------------------
// Basic construction
// --------------------------------------------------------------------------

func TestNewDefaults(t *testing.T) {
	l := New()
	assert.Equal(t, level.Trace, l.cfg.Level)
	assert.NotNil(t, l.cfg.ErrorHandler)
}

func TestNewWithConsoleAutoFormatter(t *testing.T) {
	var out testOutput
	l := New(WithOutput(&out))
	// Auto-formatter should be created.
	assert.NotNil(t, l.cfg.Formatter)
}

// Auto-select pretty when console is TTY.
func TestAutoFormatterPrettyWhenTTY(t *testing.T) {
	con := console.New(
		console.WithWriter(&bytes.Buffer{}),
		console.WithTTY(true),
	)
	l := New(WithOutput(con))
	_, ok := l.cfg.Formatter.(*pretty.PrettyFormatter)
	assert.True(t, ok, "expected PrettyFormatter when console is TTY")
}

// Auto-select text when console is not TTY.
func TestAutoFormatterTextWhenNotTTY(t *testing.T) {
	var buf bytes.Buffer
	con := console.New(console.WithWriter(&buf)) // buf → IsTTY=false
	l := New(WithOutput(con))
	_, ok := l.cfg.Formatter.(*text.TextFormatter)
	assert.True(t, ok, "expected text.Formatter when console is not TTY")
}

// Custom styles are forwarded to the pretty formatter.
func TestAutoFormatterForwardStyles(t *testing.T) {
	con := console.New(
		console.WithWriter(&bytes.Buffer{}),
		console.WithTTY(true),
		console.WithStyles(styles.Minimal()),
	)
	l := New(WithOutput(con))
	_, ok := l.cfg.Formatter.(*pretty.PrettyFormatter)
	assert.True(t, ok, "expected PrettyFormatter with custom styles")
}

// When a formatter is set explicitly, auto-selection is skipped.
func TestExplicitFormatterNotOverridden(t *testing.T) {
	var buf bytes.Buffer
	con := console.New(
		console.WithWriter(&buf),
		console.WithTTY(true),
	)
	explicit := text.New()
	l := New(WithOutput(con), WithFormatter(explicit))
	assert.Same(t, explicit, l.cfg.Formatter)
}

// WithPretty forces PrettyFormatter even without a TTY console.
func TestWithPrettyForcesPrettyFormatter(t *testing.T) {
	var out testOutput
	l := New(WithOutput(&out), WithPretty())
	_, ok := l.cfg.Formatter.(*pretty.PrettyFormatter)
	assert.True(t, ok, "expected PrettyFormatter when WithPretty is set")
}

// WithTheme is applied to the pretty formatter.
func TestWithThemeApplied(t *testing.T) {
	var out testOutput
	l := New(WithOutput(&out), WithPretty(), WithTheme(styles.Minimal()))
	_, ok := l.cfg.Formatter.(*pretty.PrettyFormatter)
	assert.True(t, ok, "expected PrettyFormatter with custom theme")
}

// WithStyles modifier is applied on top of the default theme.
func TestWithStylesModifierApplied(t *testing.T) {
	var out testOutput
	called := false
	l := New(WithOutput(&out), WithPretty(), WithStyles(func(s *styles.Styles) {
		called = true
		_ = s // mutate in real usage; here just confirm it's called
	}))
	assert.True(t, called, "WithStyles modifier must be called during New")
	_, ok := l.cfg.Formatter.(*pretty.PrettyFormatter)
	assert.True(t, ok)
}

// WithTheme overrides console.WithStyles.
func TestWithThemeOverridesConsoleStyles(t *testing.T) {
	// Both console.WithStyles and WithTheme set — WithTheme wins.
	con := console.New(
		console.WithWriter(&bytes.Buffer{}),
		console.WithTTY(true),
		console.WithStyles(styles.Minimal()),
	)
	var out testOutput
	// Adding both outputs; theme from WithTheme should win regardless.
	l := New(WithOutput(con), WithOutput(&out), WithTheme(styles.Dev()))
	assert.Equal(t, styles.Dev(), *l.cfg.prettyTheme)
	_, ok := l.cfg.Formatter.(*pretty.PrettyFormatter)
	assert.True(t, ok)
}

// --------------------------------------------------------------------------
// Level filtering
// --------------------------------------------------------------------------

func TestLevelFiltering(t *testing.T) {
	tests := []struct {
		minLevel  level.Level
		logLevel  level.Level
		wantWrite bool
	}{
		{level.Info, level.Trace, false},
		{level.Info, level.Debug, false},
		{level.Info, level.Info, true},
		{level.Info, level.Warn, true},
		{level.Info, level.Error, true},
		{level.Warn, level.Info, false},
		{level.Warn, level.Warn, true},
	}

	for _, tt := range tests {
		var out testOutput
		l := newTestLogger(&out, WithLevel(tt.minLevel))

		l.log(tt.logLevel, "msg")

		if tt.wantWrite {
			assert.NotEmpty(t, out.String(), "expected output at level %s with min %s", tt.logLevel, tt.minLevel)
		} else {
			assert.Empty(t, out.String(), "expected no output at level %s with min %s", tt.logLevel, tt.minLevel)
		}
	}
}

// --------------------------------------------------------------------------
// Log methods write correct level strings
// --------------------------------------------------------------------------

func TestLogMethods(t *testing.T) {
	tests := []struct {
		name  string
		call  func(l *Logger)
		level string
	}{
		{"Trace", func(l *Logger) { l.Trace("msg") }, "TRACE"},
		{"Debug", func(l *Logger) { l.Debug("msg") }, "DEBUG"},
		{"Info", func(l *Logger) { l.Info("msg") }, "INFO"},
		{"Warn", func(l *Logger) { l.Warn("msg") }, "WARN"},
		{"Error", func(l *Logger) { l.Error("msg") }, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out testOutput
			l := newTestLogger(&out)
			tt.call(l)
			assert.Contains(t, out.String(), tt.level)
			assert.Contains(t, out.String(), "msg")
		})
	}
}

// --------------------------------------------------------------------------
// With() — child logger field inheritance
// --------------------------------------------------------------------------

func TestWithInheritsParentFields(t *testing.T) {
	var out testOutput
	parent := newTestLogger(&out)
	child := parent.With(interfaces.String("service", "api"))

	child.Info("hello")

	assert.Contains(t, out.String(), "service=api")
	assert.Contains(t, out.String(), "hello")
}

func TestWithParentFieldsPrepeneded(t *testing.T) {
	var out testOutput
	parent := newTestLogger(&out)
	child := parent.With(interfaces.String("svc", "api"))

	child.Info("req", interfaces.String("path", "/health"))

	got := out.String()
	svcPos := bytes.Index([]byte(got), []byte("svc=api"))
	pathPos := bytes.Index([]byte(got), []byte("path=/health"))
	require.Greater(t, svcPos, 0)
	require.Greater(t, pathPos, 0)
	assert.Less(t, svcPos, pathPos, "parent field should appear before call-site field")
}

func TestWithDoesNotMutateParent(t *testing.T) {
	var out testOutput
	parent := newTestLogger(&out)
	_ = parent.With(interfaces.String("key", "val"))

	parent.Info("msg")

	assert.NotContains(t, out.String(), "key=val")
}

func TestWithChaining(t *testing.T) {
	var out testOutput
	l := newTestLogger(&out).
		With(interfaces.String("a", "1")).
		With(interfaces.String("b", "2"))

	l.Info("msg")

	got := out.String()
	assert.Contains(t, got, "a=1")
	assert.Contains(t, got, "b=2")
}

// --------------------------------------------------------------------------
// WithLevel — child logger with overridden level
// --------------------------------------------------------------------------

func TestWithLevelOverridesParent(t *testing.T) {
	var out testOutput
	// Parent allows everything; child silences DEBUG and below.
	parent := newTestLogger(&out, WithLevel(level.Trace))
	child := parent.WithLevel(level.Warn)

	child.Debug("should be filtered")
	assert.Empty(t, out.String(), "DEBUG must be filtered by child level")

	child.Warn("should pass")
	assert.Contains(t, out.String(), "should pass")
}

func TestWithLevelDoesNotMutateParent(t *testing.T) {
	var out testOutput
	parent := newTestLogger(&out, WithLevel(level.Trace))
	_ = parent.WithLevel(level.Error)

	// Parent still logs at Trace.
	parent.Debug("parent still works")
	assert.Contains(t, out.String(), "parent still works")
}

func TestWithLevelInheritsFields(t *testing.T) {
	var out testOutput
	parent := newTestLogger(&out).With(interfaces.String("svc", "api"))
	child := parent.WithLevel(level.Error)

	child.Error("oops")
	assert.Contains(t, out.String(), "svc=api")
}

func TestWithLevelAndWithCombined(t *testing.T) {
	var out testOutput
	parent := newTestLogger(&out, WithLevel(level.Trace))
	child := parent.
		WithLevel(level.Info).
		With(interfaces.String("env", "prod"))

	child.Debug("filtered")
	assert.Empty(t, out.String())

	child.Info("passes", interfaces.String("key", "val"))
	got := out.String()
	assert.Contains(t, got, "env=prod")
	assert.Contains(t, got, "key=val")
}

// --------------------------------------------------------------------------
// Error handler
// --------------------------------------------------------------------------

func TestErrorHandlerCalledOnOutputError(t *testing.T) {
	writeErr := errors.New("disk full")
	var handledErr error

	l := New(
		WithOutput(&errOutput{err: writeErr}),
		WithFormatter(text.New()),
		WithErrorHandler(func(err error) { handledErr = err }),
	)

	l.Info("msg")

	assert.Equal(t, writeErr, handledErr)
}

// --------------------------------------------------------------------------
// Caller info
// --------------------------------------------------------------------------

func TestCallerInfoDisabledByDefault(t *testing.T) {
	var out testOutput
	l := newTestLogger(&out)
	l.Info("msg")
	assert.NotContains(t, out.String(), "caller=")
}

func TestCallerInfoEnabled(t *testing.T) {
	var out testOutput
	l := newTestLogger(&out, WithCaller())
	l.Info("msg")
	assert.Contains(t, out.String(), "caller=")
	assert.Contains(t, out.String(), "logger_test.go")
}

// --------------------------------------------------------------------------
// Panic
// --------------------------------------------------------------------------

func TestPanicPanics(t *testing.T) {
	var out testOutput
	l := newTestLogger(&out)
	assert.Panics(t, func() {
		l.Panic("something went wrong")
	})
	assert.Contains(t, out.String(), "PANIC")
}

// --------------------------------------------------------------------------
// Fatal — subprocess pattern
// --------------------------------------------------------------------------

func TestFatalCallsExit(t *testing.T) {
	if os.Getenv("BE_FATAL") == "1" {
		var out testOutput
		l := newTestLogger(&out)
		l.Fatal("bye")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatalCallsExit")
	cmd.Env = append(os.Environ(), "BE_FATAL=1")
	err := cmd.Run()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitCode())
}

// --------------------------------------------------------------------------
// Concurrent writes
// --------------------------------------------------------------------------

func TestConcurrentWrites(t *testing.T) {
	var out testOutput
	l := newTestLogger(&out)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			l.Info("concurrent")
		}()
	}
	wg.Wait()

	// Each line ends with \n; count newlines as a proxy for line count.
	assert.Equal(t, goroutines, bytes.Count([]byte(out.String()), []byte("\n")))
}

// --------------------------------------------------------------------------
// Timestamp format option
// --------------------------------------------------------------------------

func TestWithTimestampFormat(t *testing.T) {
	var out testOutput
	l := New(WithOutput(&out), WithTimestampFormat("15:04:05"))
	l.Info("msg")
	got := out.String()
	// Verify the output does NOT contain the default full date format.
	assert.NotContains(t, got, "T")
}

// --------------------------------------------------------------------------
// Formatter error — routed to error handler, never panics
// --------------------------------------------------------------------------

// errFormatter always returns an error from Format.
type errFormatter struct{ err error }

func (f *errFormatter) Format(interfaces.Entry) ([]byte, error) { return nil, f.err }

func TestFormatterErrorRoutedToHandler(t *testing.T) {
	fmtErr := errors.New("formatter exploded")
	var handledErr error

	l := New(
		WithOutput(&testOutput{}),
		WithFormatter(&errFormatter{err: fmtErr}),
		WithErrorHandler(func(err error) { handledErr = err }),
	)

	assert.NotPanics(t, func() { l.Info("msg") })
	assert.Equal(t, fmtErr, handledErr)
}

// --------------------------------------------------------------------------
// Very long messages — no truncation
// --------------------------------------------------------------------------

func TestVeryLongMessageNoTruncation(t *testing.T) {
	var out testOutput
	l := newTestLogger(&out)

	// Build a 1 MiB message.
	const size = 1 << 20 // 1 MiB
	msg := string(make([]byte, size))
	l.Info(msg)

	// The output must contain every byte of the message; nothing dropped.
	got := out.String()
	assert.Contains(t, got, msg, "message must not be truncated")
}

// --------------------------------------------------------------------------
// SetLevel — concurrent level changes are race-free
// --------------------------------------------------------------------------

func TestSetLevelConcurrent(t *testing.T) {
	var out testOutput
	l := newTestLogger(&out, WithLevel(level.Trace))

	var wg sync.WaitGroup
	// Concurrent writers.
	for range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Info("concurrent")
		}()
	}
	// Concurrent level changes.
	for _, lvl := range []level.Level{level.Debug, level.Info, level.Warn, level.Trace} {
		wg.Add(1)
		go func(lv level.Level) {
			defer wg.Done()
			l.SetLevel(lv)
		}(lvl)
	}
	wg.Wait() // no race → passes with -race; no panic
}

// --------------------------------------------------------------------------
// Close — output close errors routed to error handler
// --------------------------------------------------------------------------

// errCloseOutput errors on Close.
type errCloseOutput struct {
	testOutput
	closeErr error
}

func (o *errCloseOutput) Close() error { return o.closeErr }

func TestCloseErrorRoutedToHandler(t *testing.T) {
	closeErr := errors.New("close failed")
	var handledErr error

	out := &errCloseOutput{closeErr: closeErr}
	l := New(
		WithOutput(out),
		WithFormatter(text.New()),
		WithErrorHandler(func(err error) { handledErr = err }),
	)

	l.Close()
	assert.Equal(t, closeErr, handledErr)
}

// --------------------------------------------------------------------------
// No-formatter guard — log() is a no-op when Formatter is nil
// --------------------------------------------------------------------------

func TestNoFormatterIsNoop(t *testing.T) {
	var out testOutput
	// Directly construct a Logger with a nil Formatter to exercise the guard
	// in log(). This state is unreachable via New() (auto-formatter fires),
	// but the guard must hold regardless.
	l := &Logger{
		cfg: Config{
			Outputs:      []interfaces.Output{&out},
			ErrorHandler: func(error) {},
			Clock:        clock.Real{},
		},
	}
	l.atomicLevel.Store(uint32(level.Trace))
	assert.NotPanics(t, func() { l.Info("msg") })
	assert.Empty(t, out.String())
}
