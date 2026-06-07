package json

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidjee-dev/tlog/interfaces"
	"github.com/tidjee-dev/tlog/level"
)

// fixedTime is used across all tests to produce deterministic output.
var fixedTime = time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)

func entry(lvl level.Level, msg string, fields ...interfaces.Field) interfaces.Entry {
	return interfaces.Entry{
		Timestamp: fixedTime,
		Level:     lvl,
		Message:   msg,
		Fields:    fields,
	}
}

// parseJSON parses a single JSON line from the formatter output.
func parseJSON(t *testing.T, b []byte) map[string]any {
	t.Helper()
	b = trimNewline(b)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m), "output is not valid JSON: %s", b)
	return m
}

func trimNewline(b []byte) []byte {
	if len(b) > 0 && b[len(b)-1] == '\n' {
		return b[:len(b)-1]
	}
	return b
}

// --------------------------------------------------------------------------
// Output validity — every line is valid JSON
// --------------------------------------------------------------------------

func TestOutputIsValidJSON(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "hello"))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.NotEmpty(t, m)
}

// --------------------------------------------------------------------------
// Field ordering: time, level, msg must be present
// --------------------------------------------------------------------------

func TestFieldOrdering(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "order test"))
	require.NoError(t, err)

	s := string(trimNewline(b))
	timePos := indexOf(s, `"time"`)
	levelPos := indexOf(s, `"level"`)
	msgPos := indexOf(s, `"msg"`)

	assert.Less(t, timePos, levelPos, "time must precede level")
	assert.Less(t, levelPos, msgPos, "level must precede msg")
}

func indexOf(s, sub string) int {
	for i := range s {
		if len(s)-i >= len(sub) && s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// --------------------------------------------------------------------------
// All levels
// --------------------------------------------------------------------------

func TestAllLevels(t *testing.T) {
	f := New()
	levels := []level.Level{
		level.Trace, level.Debug, level.Info,
		level.Warn, level.Error, level.Fatal, level.Panic,
	}
	for _, lvl := range levels {
		t.Run(lvl.String(), func(t *testing.T) {
			b, err := f.Format(entry(lvl, "msg"))
			require.NoError(t, err)
			m := parseJSON(t, b)
			assert.Equal(t, strings.ToLower(lvl.String()), m["level"])
		})
	}
}

// --------------------------------------------------------------------------
// Default time format is RFC3339
// --------------------------------------------------------------------------

func TestDefaultTimeFormat(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "msg"))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, "2026-05-31T12:00:00Z", m["time"])
}

func TestCustomTimeFormat(t *testing.T) {
	f := New(WithTimeFormat("2006-01-02"))
	b, err := f.Format(entry(level.Info, "msg"))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, "2026-05-31", m["time"])
}

// --------------------------------------------------------------------------
// String field
// --------------------------------------------------------------------------

func TestStringField(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "msg", interfaces.String("host", "localhost")))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, "localhost", m["host"])
}

func TestStringFieldSpecialChars(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "msg",
		interfaces.String("quote", `say "hello"`),
		interfaces.String("tab", "a\tb"),
		interfaces.String("newline", "a\nb"),
	))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, `say "hello"`, m["quote"])
	assert.Equal(t, "a\tb", m["tab"])
	assert.Equal(t, "a\nb", m["newline"])
}

// --------------------------------------------------------------------------
// Numeric fields
// --------------------------------------------------------------------------

func TestIntField(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "msg", interfaces.Int("port", 8080)))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, float64(8080), m["port"]) // JSON numbers decode as float64
}

func TestInt64Field(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "msg", interfaces.Int64("count", 1_000_000_000)))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, float64(1_000_000_000), m["count"])
}

func TestFloat64Field(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "msg", interfaces.Float64("ratio", 0.95)))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.InDelta(t, 0.95, m["ratio"], 1e-9)
}

// --------------------------------------------------------------------------
// Bool field
// --------------------------------------------------------------------------

func TestBoolField(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "msg",
		interfaces.Bool("ok", true),
		interfaces.Bool("fail", false),
	))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, true, m["ok"])
	assert.Equal(t, false, m["fail"])
}

// --------------------------------------------------------------------------
// Duration field — encoded as string
// --------------------------------------------------------------------------

func TestDurationField(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "msg",
		interfaces.Duration("latency", 322*time.Millisecond),
	))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, "322ms", m["latency"])
}

// --------------------------------------------------------------------------
// Time field — encoded as RFC3339
// --------------------------------------------------------------------------

func TestTimeField(t *testing.T) {
	f := New()
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	b, err := f.Format(entry(level.Info, "msg", interfaces.Time("at", ts)))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, "2026-01-02T03:04:05Z", m["at"])
}

// --------------------------------------------------------------------------
// Error field
// --------------------------------------------------------------------------

func TestErrField(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Error, "fail",
		interfaces.Err(errors.New("connection refused")),
	))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, "connection refused", m["error"])
}

func TestErrFieldNil(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "ok", interfaces.Err(nil)))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, "<nil>", m["error"])
}

// --------------------------------------------------------------------------
// Caller info
// --------------------------------------------------------------------------

func TestCallerInfo(t *testing.T) {
	f := New()
	e := interfaces.Entry{
		Timestamp: fixedTime,
		Level:     level.Info,
		Message:   "with caller",
		Caller:    interfaces.Caller{File: "main.go", Line: 42, Function: "main.run"},
	}
	b, err := f.Format(e)
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, "main.go:42", m["caller"])
}

func TestNoCaller(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "no caller"))
	require.NoError(t, err)
	m := parseJSON(t, b)
	_, hasCaller := m["caller"]
	assert.False(t, hasCaller)
}

// --------------------------------------------------------------------------
// Multiple fields — all present
// --------------------------------------------------------------------------

func TestMultipleFields(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "request",
		interfaces.String("method", "GET"),
		interfaces.Int("status", 200),
		interfaces.Bool("cached", true),
		interfaces.Duration("latency", 5*time.Millisecond),
	))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, "GET", m["method"])
	assert.Equal(t, float64(200), m["status"])
	assert.Equal(t, true, m["cached"])
	assert.Equal(t, "5ms", m["latency"])
}

// --------------------------------------------------------------------------
// Newline termination
// --------------------------------------------------------------------------

func TestNewlineTerminated(t *testing.T) {
	f := New()
	b, err := f.Format(entry(level.Info, "msg"))
	require.NoError(t, err)
	assert.Equal(t, byte('\n'), b[len(b)-1])
}

// --------------------------------------------------------------------------
// Very long message — no truncation
// --------------------------------------------------------------------------

func TestVeryLongMessageNoTruncation(t *testing.T) {
	f := New()
	const size = 1 << 20 // 1 MiB
	msg := string(make([]byte, size))
	b, err := f.Format(entry(level.Info, msg))
	require.NoError(t, err)
	m := parseJSON(t, b)
	assert.Equal(t, msg, m["msg"], "message must not be truncated")
}
