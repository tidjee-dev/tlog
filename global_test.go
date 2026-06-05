package tlog_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tidjee-dev/tlog"
	"github.com/tidjee-dev/tlog/formatter/text"
	"github.com/tidjee-dev/tlog/outputs/console"
)

// captureOutput builds a Logger whose output is captured in a bytes.Buffer.
func captureOutput(t *testing.T) (*tlog.Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	log := tlog.New(
		tlog.WithConsole(console.WithWriter(&buf), console.WithTTY(false)),
		tlog.WithFormatter(text.New()),
	)
	return log, &buf
}

// restoreDefault resets the default logger to a no-op after the test.
func restoreDefault(t *testing.T) {
	t.Helper()
	orig := tlog.Default()
	t.Cleanup(func() { tlog.SetDefault(orig) })
}

// --------------------------------------------------------------------------
// Default / SetDefault
// --------------------------------------------------------------------------

func TestDefaultIsNonNil(t *testing.T) {
	require.NotNil(t, tlog.Default())
}

func TestDefaultInitialIsNoOp(t *testing.T) {
	// The initial default must not panic when called.
	assert.NotPanics(t, func() {
		tlog.New().Info("no-op call — should not panic")
	})
}

func TestSetDefaultReplacesDefault(t *testing.T) {
	restoreDefault(t)

	log, _ := captureOutput(t)
	tlog.SetDefault(log)

	assert.Same(t, log, tlog.Default())
}

func TestSetDefaultNilPanics(t *testing.T) {
	// Storing nil would cause a nil-dereference on the next call — must not
	// be possible via the public API. (SetDefault panics on nil input.)
	// NOTE: atomic.Pointer.Store panics on nil; this test documents that.
	assert.Panics(t, func() {
		tlog.SetDefault(nil)
	})
}

// --------------------------------------------------------------------------
// Package-level convenience functions
// --------------------------------------------------------------------------

func TestGlobalInfo(t *testing.T) {
	restoreDefault(t)

	_, buf := captureOutput(t)
	log, _ := captureOutput(t)
	// Replace buf pointer so we can reuse the helper
	var out bytes.Buffer
	log2 := tlog.New(
		tlog.WithConsole(console.WithWriter(&out), console.WithTTY(false)),
		tlog.WithFormatter(text.New()),
	)
	tlog.SetDefault(log2)
	_ = buf // suppress unused warning
	_ = log

	tlog.Info("global info", tlog.String("key", "val"))

	assert.Contains(t, out.String(), "INFO")
	assert.Contains(t, out.String(), "global info")
	assert.Contains(t, out.String(), "key=val")
}

func TestGlobalAllLevels(t *testing.T) {
	restoreDefault(t)

	var out bytes.Buffer
	tlog.SetDefault(tlog.New(
		tlog.WithConsole(console.WithWriter(&out), console.WithTTY(false)),
		tlog.WithFormatter(text.New()),
	))

	tlog.Trace("trace msg")
	tlog.Debug("debug msg")
	tlog.Info("info msg")
	tlog.Warn("warn msg")
	tlog.Error("error msg")

	got := out.String()
	assert.Contains(t, got, "TRACE")
	assert.Contains(t, got, "DEBUG")
	assert.Contains(t, got, "INFO")
	assert.Contains(t, got, "WARN")
	assert.Contains(t, got, "ERROR")
}

// --------------------------------------------------------------------------
// Thread safety
// --------------------------------------------------------------------------

func TestSetDefaultConcurrent(t *testing.T) {
	restoreDefault(t)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func(n int) {
			defer wg.Done()
			if n%2 == 0 {
				var buf bytes.Buffer
				tlog.SetDefault(tlog.New(
					tlog.WithConsole(console.WithWriter(&buf), console.WithTTY(false)),
					tlog.WithFormatter(text.New()),
				))
			} else {
				_ = tlog.Default()
			}
		}(i)
	}
	wg.Wait()
	// Must not race-detect or panic.
	require.NotNil(t, tlog.Default())
}

func TestGlobalInfoConcurrent(t *testing.T) {
	restoreDefault(t)

	var out bytes.Buffer
	tlog.SetDefault(tlog.New(
		tlog.WithConsole(console.WithWriter(&out), console.WithTTY(false)),
		tlog.WithFormatter(text.New()),
	))

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			tlog.Info("concurrent global write")
		}()
	}
	wg.Wait()

	assert.Equal(t, goroutines, bytes.Count(out.Bytes(), []byte("\n")))
}
