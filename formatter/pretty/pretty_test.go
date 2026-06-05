package pretty_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tidjee-dev/tlog/formatter/pretty"
	"github.com/tidjee-dev/tlog/interfaces"
	"github.com/tidjee-dev/tlog/level"
	"github.com/tidjee-dev/tlog/styles"
)

var fixedTime = time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)

func entry(lvl level.Level, msg string, fields ...interfaces.Field) interfaces.Entry {
	return interfaces.Entry{
		Timestamp: fixedTime,
		Level:     lvl,
		Message:   msg,
		Fields:    fields,
	}
}

// ── Plain-text fallback (isTTY=false) ────────────────────────────────────────

func TestPlain_AllLevels(t *testing.T) {
	f := pretty.New(pretty.WithTTY(false))

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

func TestPlain_WithFields(t *testing.T) {
	f := pretty.New(pretty.WithTTY(false))
	e := entry(level.Info, "request",
		interfaces.String("method", "GET"),
		interfaces.Int("status", 200),
	)
	got, err := f.Format(e)
	require.NoError(t, err)
	assert.Equal(t, "2026-05-31T12:00:00.000 INFO  request  method=GET  status=200\n", string(got))
}

func TestPlain_WithCaller(t *testing.T) {
	f := pretty.New(pretty.WithTTY(false))
	e := interfaces.Entry{
		Timestamp: fixedTime,
		Level:     level.Info,
		Message:   "with caller",
		Caller:    interfaces.Caller{File: "main.go", Line: 42},
	}
	got, err := f.Format(e)
	require.NoError(t, err)
	assert.Equal(t, "2026-05-31T12:00:00.000 INFO  with caller  caller=main.go:42\n", string(got))
}

func TestPlain_QuotedValues(t *testing.T) {
	f := pretty.New(pretty.WithTTY(false))
	e := entry(level.Warn, "msg",
		interfaces.String("msg", "hello world"),
		interfaces.Err(errors.New("bad error")),
	)
	got, err := f.Format(e)
	require.NoError(t, err)
	assert.Contains(t, string(got), `msg="hello world"`)
	assert.Contains(t, string(got), `error="bad error"`)
}

func TestPlain_CustomTimestamp(t *testing.T) {
	f := pretty.New(pretty.WithTTY(false), pretty.WithTimestampFormat("15:04:05"))
	got, err := f.Format(entry(level.Info, "ts"))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(got), "12:00:00 "))
}

// ── Styled output (isTTY=true) ───────────────────────────────────────────────

func TestStyled_ContainsMessage(t *testing.T) {
	f := pretty.New() // isTTY=true by default
	got, err := f.Format(entry(level.Info, "server started"))
	require.NoError(t, err)
	assert.Contains(t, string(got), "server started")
	assert.True(t, strings.HasSuffix(string(got), "\n"))
}

func TestStyled_AllLevels_ContainLevelString(t *testing.T) {
	f := pretty.New()
	levels := []level.Level{
		level.Trace, level.Debug, level.Info,
		level.Warn, level.Error, level.Fatal, level.Panic,
	}
	for _, lvl := range levels {
		t.Run(lvl.String(), func(t *testing.T) {
			got, err := f.Format(entry(lvl, "msg"))
			require.NoError(t, err)
			assert.Contains(t, string(got), lvl.String())
		})
	}
}

func TestStyled_ContainsFieldKeyAndValue(t *testing.T) {
	f := pretty.New()
	e := entry(level.Info, "req",
		interfaces.String("method", "GET"),
		interfaces.Int("status", 200),
	)
	got, err := f.Format(e)
	require.NoError(t, err)
	assert.Contains(t, string(got), "method")
	assert.Contains(t, string(got), "GET")
	assert.Contains(t, string(got), "status")
	assert.Contains(t, string(got), "200")
}

func TestStyled_ContainsCaller(t *testing.T) {
	f := pretty.New()
	e := interfaces.Entry{
		Timestamp: fixedTime,
		Level:     level.Info,
		Message:   "msg",
		Caller:    interfaces.Caller{File: "app.go", Line: 99},
	}
	got, err := f.Format(e)
	require.NoError(t, err)
	assert.Contains(t, string(got), "app.go")
	assert.Contains(t, string(got), "99")
}

func TestStyled_NoColorTheme_MatchesPlain(t *testing.T) {
	// NoColor theme with isTTY=true should produce output identical to
	// plain-text (no ANSI escapes) but still go through the styled path.
	styled := pretty.New(
		pretty.WithStyles(styles.NoColor()),
		pretty.WithTTY(true),
	)
	plain := pretty.New(pretty.WithTTY(false))

	e := entry(level.Info, "hello", interfaces.String("key", "val"))

	styledOut, err := styled.Format(e)
	require.NoError(t, err)
	plainOut, err := plain.Format(e)
	require.NoError(t, err)

	assert.Equal(t, string(plainOut), string(styledOut))
}

func TestStyled_CustomTheme(t *testing.T) {
	f := pretty.New(pretty.WithStyles(styles.Dev()))
	got, err := f.Format(entry(level.Error, "boom"))
	require.NoError(t, err)
	assert.Contains(t, string(got), "boom")
}
