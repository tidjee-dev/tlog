package interfaces

import (
	"time"

	"github.com/tidjee-dev/tlog/level"
)

// Entry is an immutable log event. It is built once by the logger and passed
// to formatters and outputs. No public mutators are provided.
type Entry struct {
	Timestamp time.Time
	Level     level.Level
	Message   string
	Fields    []Field
	Caller    Caller
}

// Caller holds optional source location information for a log entry.
// It is populated only when the logger is configured with WithCaller().
type Caller struct {
	File     string
	Line     int
	Function string
}

// IsZero reports whether the Caller is empty (caller info not enabled).
func (c Caller) IsZero() bool {
	return c.File == "" && c.Line == 0 && c.Function == ""
}
