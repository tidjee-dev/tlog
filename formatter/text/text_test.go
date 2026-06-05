package text

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidjee-dev/tlog/interfaces"
	"github.com/tidjee-dev/tlog/level"
)

// fixedTime is a stable timestamp used across all tests.
var fixedTime = time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)

func entry(lvl level.Level, msg string, fields ...interfaces.Field) interfaces.Entry {
	return interfaces.Entry{
		Timestamp: fixedTime,
		Level:     lvl,
		Message:   msg,
		Fields:    fields,
	}
}

func TestFormatAllLevels(t *testing.T) {
	f := New()

	tests := []struct {
		lvl  level.Level
		want string
	}{
		{level.Trace, "2026-05-31T12:00:00.000 TRACE hello\n"},
		{level.Debug, "2026-05-31T12:00:00.000 DEBUG hello\n"},
		{level.Info, "2026-05-31T12:00:00.000 INFO  hello\n"},
		{level.Warn, "2026-05-31T12:00:00.000 WARN  hello\n"},
		{level.Error, "2026-05-31T12:00:00.000 ERROR hello\n"},
		{level.Fatal, "2026-05-31T12:00:00.000 FATAL hello\n"},
		{level.Panic, "2026-05-31T12:00:00.000 PANIC hello\n"},
	}

	for _, tt := range tests {
		t.Run(tt.lvl.String(), func(t *testing.T) {
			got, err := f.Format(entry(tt.lvl, "hello"))
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestFormatNoFields(t *testing.T) {
	f := New()
	got, err := f.Format(entry(level.Info, "server started"))
	require.NoError(t, err)
	assert.Equal(t, "2026-05-31T12:00:00.000 INFO  server started\n", string(got))
}

func TestFormatWithFields(t *testing.T) {
	f := New()
	got, err := f.Format(entry(level.Info, "request",
		interfaces.String("method", "GET"),
		interfaces.Int("status", 200),
		interfaces.Bool("cached", true),
	))
	require.NoError(t, err)
	assert.Equal(t, "2026-05-31T12:00:00.000 INFO  request  method=GET  status=200  cached=true\n", string(got))
}

func TestFormatStringQuoting(t *testing.T) {
	f := New()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"plain", "localhost", "2026-05-31T12:00:00.000 INFO  msg  host=localhost\n"},
		{"with space", "hello world", "2026-05-31T12:00:00.000 INFO  msg  host=\"hello world\"\n"},
		{"empty", "", "2026-05-31T12:00:00.000 INFO  msg  host=\"\"\n"},
		{"with equals", "a=b", "2026-05-31T12:00:00.000 INFO  msg  host=\"a=b\"\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.Format(entry(level.Info, "msg", interfaces.String("host", tt.value)))
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestFormatWithError(t *testing.T) {
	f := New()

	t.Run("with error", func(t *testing.T) {
		got, err := f.Format(entry(level.Error, "db failed",
			interfaces.Err(errors.New("connection refused")),
		))
		require.NoError(t, err)
		assert.Equal(t, "2026-05-31T12:00:00.000 ERROR db failed  error=\"connection refused\"\n", string(got))
	})

	t.Run("nil error", func(t *testing.T) {
		got, err := f.Format(entry(level.Info, "ok", interfaces.Err(nil)))
		require.NoError(t, err)
		assert.Equal(t, "2026-05-31T12:00:00.000 INFO  ok  error=<nil>\n", string(got))
	})
}

func TestFormatWithDuration(t *testing.T) {
	f := New()
	got, err := f.Format(entry(level.Info, "done",
		interfaces.Duration("latency", 322*time.Millisecond),
	))
	require.NoError(t, err)
	assert.Equal(t, "2026-05-31T12:00:00.000 INFO  done  latency=322ms\n", string(got))
}

func TestFormatWithTime(t *testing.T) {
	f := New()
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	got, err := f.Format(entry(level.Info, "event", interfaces.Time("at", ts)))
	require.NoError(t, err)
	assert.Equal(t, "2026-05-31T12:00:00.000 INFO  event  at=2026-01-02T03:04:05Z\n", string(got))
}

func TestFormatWithCaller(t *testing.T) {
	f := New()
	e := interfaces.Entry{
		Timestamp: fixedTime,
		Level:     level.Info,
		Message:   "trace me",
		Caller:    interfaces.Caller{File: "main.go", Line: 42, Function: "main.run"},
	}
	got, err := f.Format(e)
	require.NoError(t, err)
	assert.Equal(t, "2026-05-31T12:00:00.000 INFO  trace me  caller=main.go:42\n", string(got))
}

func TestFormatWithFieldsAndCaller(t *testing.T) {
	f := New()
	e := interfaces.Entry{
		Timestamp: fixedTime,
		Level:     level.Warn,
		Message:   "slow",
		Fields:    []interfaces.Field{interfaces.Duration("latency", 500*time.Millisecond)},
		Caller:    interfaces.Caller{File: "server.go", Line: 99},
	}
	got, err := f.Format(e)
	require.NoError(t, err)
	assert.Equal(t, "2026-05-31T12:00:00.000 WARN  slow  latency=500ms  caller=server.go:99\n", string(got))
}

func TestFormatCustomTimestampFormat(t *testing.T) {
	f := New(WithTimestampFormat("15:04:05"))
	got, err := f.Format(entry(level.Info, "hello"))
	require.NoError(t, err)
	assert.Equal(t, "12:00:00 INFO  hello\n", string(got))
}

// --------------------------------------------------------------------------
// Very long message — no truncation
// --------------------------------------------------------------------------

func TestVeryLongMessageNoTruncation(t *testing.T) {
	f := New()
	const size = 1 << 20 // 1 MiB
	msg := string(make([]byte, size))
	got, err := f.Format(entry(level.Info, msg))
	require.NoError(t, err)
	assert.Contains(t, string(got), msg, "message must not be truncated")
}
