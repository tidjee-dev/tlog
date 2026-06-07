// Package pretty provides a lipgloss-styled formatter for human-readable
// console output. It produces aligned columns:
//
//	<timestamp>  <LEVEL>  <message>  <key>=<value> ...  <caller>
//
// When the output is not a TTY the formatter falls back to plain text
// (identical output to formatter/text).
package pretty

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/tidjee-dev/tlog/interfaces"
	"github.com/tidjee-dev/tlog/internal/buffer"
	"github.com/tidjee-dev/tlog/level"
	"github.com/tidjee-dev/tlog/styles"
)

// PrettyFormatter formats log entries with lipgloss styling.
type PrettyFormatter struct {
	styles          styles.Styles
	timestampFormat string
	isTTY           bool
}

// padLevel left-pads s to exactly 5 characters without allocation.
func padLevel(s string) string {
	const spaces = "     "
	if len(s) >= 5 {
		return s
	}
	return s + spaces[:5-len(s)]
}

// Option configures a PrettyFormatter.
type Option func(*PrettyFormatter)

// WithStyles replaces the full Styles set.
func WithStyles(s styles.Styles) Option {
	return func(f *PrettyFormatter) {
		f.styles = s
	}
}

// WithTimestampFormat sets the Go time layout for the timestamp column.
func WithTimestampFormat(format string) Option {
	return func(f *PrettyFormatter) {
		f.timestampFormat = format
	}
}

// WithTTY overrides TTY detection. Pass true to force styled output,
// false to force plain-text output.
func WithTTY(tty bool) Option {
	return func(f *PrettyFormatter) {
		f.isTTY = tty
	}
}

// New returns a PrettyFormatter. By default it uses the Dev theme,
// the standard timestamp format, and assumes a TTY.
// Pass WithTTY(false) to get plain-text output from this formatter.
func New(opts ...Option) *PrettyFormatter {
	f := &PrettyFormatter{
		styles:          styles.Dev(),
		timestampFormat: "2006-01-02T15:04:05.000",
		isTTY:           true,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// Format implements interfaces.Formatter.
// Produces a newline-terminated styled line (or plain text when not TTY).
func (f *PrettyFormatter) Format(entry interfaces.Entry) ([]byte, error) {
	if !f.isTTY {
		return f.formatPlain(entry)
	}
	return f.formatStyled(entry)
}

// formatStyled renders a lipgloss-coloured line.
func (f *PrettyFormatter) formatStyled(entry interfaces.Entry) ([]byte, error) {
	buf := buffer.Get()
	defer buffer.Put(buf)

	s := f.styles

	// Timestamp
	buf.WriteString(s.Timestamp.Render(entry.Timestamp.Format(f.timestampFormat)))
	buf.WriteByte(' ')

	// Level (padded to 5)
	levelStyle := f.levelStyle(entry.Level)
	buf.WriteString(levelStyle.Render(padLevel(entry.Level.String())))
	buf.WriteByte(' ')

	// Message
	buf.WriteString(s.Message.Render(entry.Message))

	// Fields: key=value
	for _, field := range entry.Fields {
		buf.WriteString("  ")
		buf.WriteString(s.Key.Render(field.Key))
		buf.WriteByte('=')
		buf.WriteString(s.Value.Render(formatValue(field)))
	}

	// Caller
	if !entry.Caller.IsZero() {
		buf.WriteString("  ")
		callerStr := entry.Caller.File + ":" + strconv.Itoa(entry.Caller.Line)
		buf.WriteString(s.Caller.Render(callerStr))
	}

	buf.WriteByte('\n')
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

// formatPlain renders a plain-text line (identical to formatter/text output).
func (f *PrettyFormatter) formatPlain(entry interfaces.Entry) ([]byte, error) {
	buf := buffer.Get()
	defer buffer.Put(buf)

	buf.WriteString(entry.Timestamp.Format(f.timestampFormat))
	buf.WriteByte(' ')
	buf.WriteString(padLevel(entry.Level.String()))
	buf.WriteByte(' ')
	buf.WriteString(entry.Message)

	for _, field := range entry.Fields {
		buf.WriteString("  ")
		buf.WriteString(field.Key)
		buf.WriteByte('=')
		buf.WriteString(formatValue(field))
	}

	if !entry.Caller.IsZero() {
		buf.WriteString("  caller=")
		buf.WriteString(entry.Caller.File)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(entry.Caller.Line))
	}

	buf.WriteByte('\n')
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

// levelStyle returns the lipgloss.Style for the given level.
func (f *PrettyFormatter) levelStyle(l level.Level) interface{ Render(...string) string } {
	switch l {
	case level.Trace:
		return f.styles.LevelTrace
	case level.Debug:
		return f.styles.LevelDebug
	case level.Info:
		return f.styles.LevelInfo
	case level.Warn:
		return f.styles.LevelWarn
	case level.Error:
		return f.styles.LevelError
	case level.Fatal:
		return f.styles.LevelFatal
	case level.Panic:
		return f.styles.LevelPanic
	default:
		return f.styles.LevelInfo
	}
}

// formatValue renders a Field's value as a plain string.
func formatValue(f interfaces.Field) string {
	switch f.Type {
	case interfaces.StringType:
		s, _ := f.Value.(string)
		if needsQuoting(s) {
			return strconv.Quote(s)
		}
		return s
	case interfaces.IntType:
		v, _ := f.Value.(int)
		return strconv.Itoa(v)
	case interfaces.Int64Type:
		v, _ := f.Value.(int64)
		return strconv.FormatInt(v, 10)
	case interfaces.Float64Type:
		v, _ := f.Value.(float64)
		return strconv.FormatFloat(v, 'f', -1, 64)
	case interfaces.BoolType:
		v, _ := f.Value.(bool)
		return strconv.FormatBool(v)
	case interfaces.DurationType:
		v, ok := f.Value.(time.Duration)
		if !ok {
			return "<invalid-duration>"
		}

		switch {
		case v == 0:
			return "0s"

		case v < time.Microsecond:
			return v.Round(time.Nanosecond).String()

		case v < time.Millisecond:
			return v.Round(time.Microsecond).String()

		case v < time.Second:
			return v.Round(100 * time.Microsecond).String()

		case v < time.Minute:
			return v.Round(time.Millisecond).String()

		default:
			return v.Round(time.Second).String()
		}
	case interfaces.TimeType:
		t, _ := f.Value.(time.Time)
		return t.Format(time.RFC3339)
	case interfaces.ErrorType:
		err, ok := f.Value.(error)
		if !ok || err == nil {
			return "<nil>"
		}
		s := err.Error()
		if needsQuoting(s) {
			return strconv.Quote(s)
		}
		return s
	case interfaces.AnyType:
		if f.Value == nil {
			return "<nil>"
		}
		if b, err := json.Marshal(f.Value); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", f.Value)
	default:
		return "<unsupported>"
	}
}

// needsQuoting reports whether s must be wrapped in double quotes.
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	for _, c := range s {
		if c == ' ' || c == '"' || c == '\\' || c == '=' || c == '\n' || c == '\t' {
			return true
		}
	}
	return false
}
