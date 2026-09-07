// Package clock provides a mockable time source for use in logging internals.
// Production code uses Real, which delegates to time.Now().
// Tests can inject Mock to control time deterministically.
package clock

import (
	"sync"
	"time"
)

// Clock is a time provider.
type Clock interface {
	Now() time.Time
}

// Real is a Clock backed by time.Now.
type Real struct{}

// Now returns the current wall-clock time.
func (Real) Now() time.Time { return time.Now() }

// Mock is a Clock whose time can be set explicitly.
// It is safe for concurrent use.
type Mock struct {
	mu sync.RWMutex
	t  time.Time
}

// NewMock returns a Mock clock initialised to t.
func NewMock(t time.Time) *Mock { return &Mock{t: t} }

// Set advances the mock clock to t.
func (m *Mock) Set(t time.Time) {
	m.mu.Lock()
	m.t = t
	m.mu.Unlock()
}

// Now returns the mock's current time.
func (m *Mock) Now() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.t
}
