package level

import (
	"fmt"
	"strings"
)

// Level represents a log severity level.
type Level uint8

const (
	Trace Level = iota
	Debug
	Info
	Warn
	Error
	Fatal
	Panic
)

var levelNames = [...]string{
	Trace: "TRACE",
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
	Fatal: "FATAL",
	Panic: "PANIC",
}

// String returns the uppercase string representation of the level.
func (l Level) String() string {
	if int(l) < len(levelNames) {
		return levelNames[l]
	}
	return fmt.Sprintf("LEVEL(%d)", l)
}

// Parse returns the Level corresponding to the given string (case-insensitive).
// Returns an error if the string does not match any known level.
func Parse(s string) (Level, error) {
	switch strings.ToUpper(s) {
	case "TRACE":
		return Trace, nil
	case "DEBUG":
		return Debug, nil
	case "INFO":
		return Info, nil
	case "WARN":
		return Warn, nil
	case "ERROR":
		return Error, nil
	case "FATAL":
		return Fatal, nil
	case "PANIC":
		return Panic, nil
	default:
		return 0, fmt.Errorf("level: unknown level %q", s)
	}
}

// MustParse returns the Level for the given string (case-insensitive).
// Panics if the string does not match any known level.
func MustParse(s string) Level {
	l, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return l
}
