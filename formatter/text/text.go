package text

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tidjee-dev/tlog/interfaces"
	"github.com/tidjee-dev/tlog/internal/buffer"
)

// TextFormatter formats log entries as human-readable plain text.
// Output format: <timestamp> <LEVEL> <message>  [key=value ...] [caller=file:line]
type TextFormatter struct {
	timestampFormat string
}

// padLevel right-pads s with spaces to 5 characters (allocates for short
// levels such as INFO/WARN; TRACE/DEBUG/ERROR/FATAL/PANIC already fill it).
func padLevel(s string) string {
	const spaces = "     "
	if len(s) >= 5 {
		return s
	}
	return s + spaces[:5-len(s)]
}

// Option configures a TextFormatter.
type Option func(*TextFormatter)

// WithTimestampFormat sets the Go time layout used for the timestamp field.
func WithTimestampFormat(format string) Option {
	return func(tf *TextFormatter) {
		tf.timestampFormat = format
	}
}

// New returns a TextFormatter with the given options applied.
// Default timestamp format: "2006-01-02T15:04:05.000".
func New(opts ...Option) *TextFormatter {
	tf := &TextFormatter{
		timestampFormat: "2006-01-02T15:04:05.000",
	}
	for _, opt := range opts {
		opt(tf)
	}
	return tf
}

// Format implements interfaces.Formatter.
// Produces a newline-terminated plain-text line.
func (tf *TextFormatter) Format(entry interfaces.Entry) ([]byte, error) {
	buf := buffer.Get()
	defer buffer.Put(buf)

	buf.WriteString(entry.Timestamp.Format(tf.timestampFormat))
	buf.WriteByte(' ')
	buf.WriteString(padLevel(entry.Level.String()))
	buf.WriteByte(' ')
	buf.WriteString(formatMessage(entry.Message))

	for _, f := range entry.Fields {
		buf.WriteString("  ")
		buf.WriteString(f.Key)
		buf.WriteByte('=')
		buf.WriteString(formatValue(f))
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

// formatMessage renders the message so one entry stays on one line:
// messages containing \n or \r are double-quoted with escapes.
func formatMessage(msg string) string {
	if strings.ContainsAny(msg, "\r\n") {
		return strconv.Quote(msg)
	}
	return msg
}

// formatValue renders a Field's value as a string, quoting when necessary.
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

		// Compare on magnitude so negative durations round like positives.
		mag := v
		if mag < 0 {
			mag = -mag
		}
		switch {
		case v == 0:
			return "0s"

		case mag < time.Microsecond:
			return v.Round(time.Nanosecond).String()

		case mag < time.Millisecond:
			return v.Round(time.Microsecond).String()

		case mag < time.Second:
			return v.Round(100 * time.Microsecond).String()

		case mag < time.Minute:
			return v.Round(time.Millisecond).String()

		default:
			return v.Round(time.Second).String()
		}
	case interfaces.TimeType:
		t, ok := f.Value.(time.Time)
		if !ok {
			return "<invalid-time>"
		}
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

// needsQuoting reports whether s must be wrapped in double quotes in the output.
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	for _, c := range s {
		if c == ' ' || c == '"' || c == '\\' || c == '=' || c == '\n' || c == '\r' || c == '\t' {
			return true
		}
	}
	return false
}
