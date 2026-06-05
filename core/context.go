package core

import (
	"context"

	"github.com/tidjee-dev/tlog/interfaces"
)

// contextKey is the unexported key type used to store a Logger in a context.
// Using a private type prevents collisions with other packages.
type contextKey struct{}

// NewContext returns a new context carrying log. Retrieve it with FromContext.
func NewContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

// FromContext returns the Logger stored in ctx by NewContext.
// If no logger is present, a no-op logger (no outputs, no formatter) is
// returned so callers never need to nil-check.
func FromContext(ctx context.Context) *Logger {
	if log, ok := ctx.Value(contextKey{}).(*Logger); ok && log != nil {
		return log
	}
	return New()
}

// WithContext returns a copy of ctx whose embedded logger has fields appended.
// It is a shorthand for:
//
//	core.NewContext(ctx, core.FromContext(ctx).With(fields...))
func WithContext(ctx context.Context, fields ...interfaces.Field) context.Context {
	return NewContext(ctx, FromContext(ctx).With(fields...))
}
