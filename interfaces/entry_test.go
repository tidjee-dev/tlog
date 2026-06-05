package interfaces

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidjee-dev/tlog/level"
)

func TestEntryFields(t *testing.T) {
	now := time.Now()
	e := Entry{
		Timestamp: now,
		Level:     level.Info,
		Message:   "hello world",
		Fields:    []Field{String("key", "val")},
	}

	assert.Equal(t, now, e.Timestamp)
	assert.Equal(t, level.Info, e.Level)
	assert.Equal(t, "hello world", e.Message)
	assert.Len(t, e.Fields, 1)
	assert.Equal(t, "key", e.Fields[0].Key)
	assert.True(t, e.Caller.IsZero())
}

func TestCallerIsZero(t *testing.T) {
	assert.True(t, Caller{}.IsZero())
	assert.False(t, Caller{File: "main.go", Line: 42}.IsZero())
}
