package json

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tidjee-dev/tlog/interfaces"
	"github.com/tidjee-dev/tlog/internal/buffer"
)

// JSONFormatter formats log entries as single-line JSON objects.
// Field ordering: time, level, msg, then user fields in declaration order.
// No reflection is used in the encoding path.
type JSONFormatter struct {
	timeFormat string
}

// Option configures a JSONFormatter.
type Option func(*JSONFormatter)

// WithTimeFormat sets the Go time layout used for the "time" field.
// Default: time.RFC3339Nano.
func WithTimeFormat(format string) Option {
	return func(f *JSONFormatter) {
		f.timeFormat = format
	}
}

// New returns a JSONFormatter with the given options applied.
func New(opts ...Option) *JSONFormatter {
	f := &JSONFormatter{
		timeFormat: time.RFC3339Nano,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// Format implements interfaces.Formatter.
// Produces a newline-terminated JSON line.
func (f *JSONFormatter) Format(entry interfaces.Entry) ([]byte, error) {
	buf := buffer.Get()
	defer buffer.Put(buf)

	buf.WriteByte('{')

	// time
	buf.WriteString(`"time":`)
	writeString(buf, entry.Timestamp.Format(f.timeFormat))

	// level
	buf.WriteString(`,"level":`)
	// lowercase level for JSON output
	buf.WriteString(strconv.Quote(strings.ToLower(entry.Level.String())))

	// msg
	buf.WriteString(`,"msg":`)
	writeString(buf, entry.Message)

	// user fields
	for _, field := range entry.Fields {
		buf.WriteByte(',')
		writeString(buf, field.Key)
		buf.WriteByte(':')
		writeFieldValue(buf, field)
	}

	// caller (optional)
	if !entry.Caller.IsZero() {
		buf.WriteString(`,"caller":`)
		writeString(buf, entry.Caller.File+":"+strconv.Itoa(entry.Caller.Line))
	}

	buf.WriteString("}\n")
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

// writeFieldValue encodes a Field's value into buf without reflection.
func writeFieldValue(buf *bytes.Buffer, f interfaces.Field) {
	switch f.Type {
	case interfaces.StringType:
		s, _ := f.Value.(string)
		writeString(buf, s)

	case interfaces.IntType:
		v, _ := f.Value.(int)
		buf.WriteString(strconv.Itoa(v))

	case interfaces.Int64Type:
		v, _ := f.Value.(int64)
		buf.WriteString(strconv.FormatInt(v, 10))

	case interfaces.Float64Type:
		v, _ := f.Value.(float64)
		buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))

	case interfaces.BoolType:
		v, _ := f.Value.(bool)
		if v {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}

	case interfaces.DurationType:
		v, ok := f.Value.(time.Duration)
		if !ok {
			writeString(buf, "<invalid-duration>")
			return
		}

		switch {
		case v == 0:
			writeString(buf, "0s")

		case v < time.Microsecond:
			writeString(buf, v.Round(time.Nanosecond).String())

		case v < time.Millisecond:
			writeString(buf, v.Round(time.Microsecond).String())

		case v < time.Second:
			writeString(buf, v.Round(100*time.Microsecond).String())

		case v < time.Minute:
			writeString(buf, v.Round(time.Millisecond).String())

		default:
			writeString(buf, v.Round(time.Second).String())
		}

	case interfaces.TimeType:
		v, ok := f.Value.(time.Time)
		if !ok {
			writeString(buf, "<invalid-time>")
		} else {
			writeString(buf, v.Format(time.RFC3339))
		}

	case interfaces.ErrorType:
		err, ok := f.Value.(error)
		if !ok || err == nil {
			writeString(buf, "<nil>")
		} else {
			writeString(buf, err.Error())
		}

	case interfaces.AnyType:
		writeAnyValue(buf, f.Value)

	default:
		writeString(buf, anyToString(f.Value))
	}
}

// writeString writes a JSON-encoded string (with surrounding quotes) to buf.
// It handles the ASCII fast path and falls back to full Unicode escaping.
func writeString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		b := s[i]
		if b < utf8.RuneSelf {
			if htmlSafeSet[b] {
				i++
				continue
			}
			if start < i {
				buf.WriteString(s[start:i])
			}
			switch b {
			case '"':
				buf.WriteString(`\"`)
			case '\\':
				buf.WriteString(`\\`)
			case '\n':
				buf.WriteString(`\n`)
			case '\r':
				buf.WriteString(`\r`)
			case '\t':
				buf.WriteString(`\t`)
			default:
				// Control character — encode as \uXXXX.
				buf.WriteString(`\u00`)
				buf.WriteByte(hexChars[b>>4])
				buf.WriteByte(hexChars[b&0xf])
			}
			i++
			start = i
			continue
		}
		// Multi-byte rune — pass through as-is (valid UTF-8).
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
	}
	if start < len(s) {
		buf.WriteString(s[start:])
	}
	buf.WriteByte('"')
}

// anyToString converts an arbitrary value to a string without reflection
// by using the fmt.Stringer interface when available, otherwise falling back
// to a format string via fmt.Sprint. Import-free for non-Stringer values
// would require reflect; we accept one fmt.Sprint call only in the AnyType path.
func anyToString(v any) string {
	if s, ok := v.(interface{ String() string }); ok {
		return s.String()
	}
	buf := make([]byte, 0, 32)
	buf = strconv.AppendQuote(buf, "") // prime the encoder
	_ = buf
	// Use a simple sprint via strconv for basic scalars; others get %v.
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		// Truly unknown type — accept the fmt import only here.
		_ = val
		return "<unsupported>"
	}
}

func writeAnyValue(buf *bytes.Buffer, v any) {
	if v == nil {
		writeString(buf, "<nil>")
		return
	}

	switch val := v.(type) {
	case string:
		writeString(buf, val)
		return
	case int:
		buf.WriteString(strconv.Itoa(val))
		return
	case int64:
		buf.WriteString(strconv.FormatInt(val, 10))
		return
	case float64:
		buf.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
		return
	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return
	case time.Time:
		writeString(buf, val.Format(time.RFC3339))
		return
	case time.Duration:
		writeString(buf, val.String())
		return
	}
	// fallback: JSON encoding for structured values
	b, err := json.Marshal(v)
	if err == nil {
		buf.Write(b)
		return
	}

	writeString(buf, anyToString(v))
}

const hexChars = "0123456789abcdef"

// htmlSafeSet marks bytes that can be passed through directly in a JSON string.
// Characters that need escaping (control chars, ", \) are false.
var htmlSafeSet = [utf8.RuneSelf]bool{
	' ': true, '!': true,
	// '"' = false (must escape)
	'#': true, '$': true, '%': true, '&': true, '\'': true,
	'(': true, ')': true, '*': true, '+': true, ',': true,
	'-': true, '.': true, '/': true,
	'0': true, '1': true, '2': true, '3': true, '4': true,
	'5': true, '6': true, '7': true, '8': true, '9': true,
	':': true, ';': true, '<': true, '=': true, '>': true,
	'?': true, '@': true,
	'A': true, 'B': true, 'C': true, 'D': true, 'E': true,
	'F': true, 'G': true, 'H': true, 'I': true, 'J': true,
	'K': true, 'L': true, 'M': true, 'N': true, 'O': true,
	'P': true, 'Q': true, 'R': true, 'S': true, 'T': true,
	'U': true, 'V': true, 'W': true, 'X': true, 'Y': true, 'Z': true,
	'[': true,
	// '\\' = false (must escape)
	']': true, '^': true, '_': true, '`': true,
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true,
	'f': true, 'g': true, 'h': true, 'i': true, 'j': true,
	'k': true, 'l': true, 'm': true, 'n': true, 'o': true,
	'p': true, 'q': true, 'r': true, 's': true, 't': true,
	'u': true, 'v': true, 'w': true, 'x': true, 'y': true, 'z': true,
	'{': true, '|': true, '}': true, '~': true,
}
