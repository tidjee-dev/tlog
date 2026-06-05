package clock_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidjee-dev/tlog/internal/clock"
)

func TestReal_Now(t *testing.T) {
	before := time.Now()
	got := clock.Real{}.Now()
	after := time.Now()
	assert.True(t, !got.Before(before) && !got.After(after),
		"Real.Now() should be between before and after")
}

func TestMock_Now(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	m := clock.NewMock(fixed)
	assert.Equal(t, fixed, m.Now())
}

func TestMock_Set(t *testing.T) {
	m := clock.NewMock(time.Now())
	next := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	m.Set(next)
	assert.Equal(t, next, m.Now())
}

func TestClockInterface(t *testing.T) {
	// Both Real and *Mock must satisfy Clock.
	var _ clock.Clock = clock.Real{}
	var _ clock.Clock = (*clock.Mock)(nil)
}
