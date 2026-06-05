package core

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidjee-dev/tlog/formatter/text"
	"github.com/tidjee-dev/tlog/interfaces"
)

// --------------------------------------------------------------------------
// NewContext / FromContext round-trip
// --------------------------------------------------------------------------

func TestNewContextFromContextRoundTrip(t *testing.T) {
	var out testOutput
	log := newTestLogger(&out)

	ctx := NewContext(context.Background(), log)
	got := FromContext(ctx)

	require.NotNil(t, got)
	got.Info("round-trip")
	assert.Contains(t, out.String(), "round-trip")
}

func TestFromContextMissingReturnsNoOp(t *testing.T) {
	// No logger stored — should return a non-nil no-op logger, never panic.
	log := FromContext(context.Background())
	require.NotNil(t, log)
	// Logging on the no-op must not panic (no formatter, no outputs).
	assert.NotPanics(t, func() {
		log.Info("this is silently dropped")
	})
}

func TestFromContextNilValueReturnsNoOp(t *testing.T) {
	// Explicitly store nil — should still return a non-nil logger.
	ctx := context.WithValue(context.Background(), contextKey{}, (*Logger)(nil))
	log := FromContext(ctx)
	require.NotNil(t, log)
}

// --------------------------------------------------------------------------
// WithContext — field attachment
// --------------------------------------------------------------------------

func TestWithContextAppendsFields(t *testing.T) {
	var out testOutput
	log := New(WithOutput(&out), WithFormatter(text.New()))

	ctx := NewContext(context.Background(), log)
	ctx = WithContext(ctx, interfaces.String("request_id", "abc-123"))

	FromContext(ctx).Info("handling request")

	assert.Contains(t, out.String(), "request_id=abc-123")
	assert.Contains(t, out.String(), "handling request")
}

func TestWithContextDoesNotMutateOriginalLogger(t *testing.T) {
	var out testOutput
	log := New(WithOutput(&out), WithFormatter(text.New()))

	ctx := NewContext(context.Background(), log)
	_ = WithContext(ctx, interfaces.String("key", "value"))

	// The original logger (not from the mutated context) must be field-free.
	log.Info("original")
	assert.NotContains(t, out.String(), "key=value")
}

func TestWithContextChaining(t *testing.T) {
	var out testOutput
	log := New(WithOutput(&out), WithFormatter(text.New()))

	ctx := NewContext(context.Background(), log)
	ctx = WithContext(ctx, interfaces.String("service", "api"))
	ctx = WithContext(ctx, interfaces.String("version", "v2"))

	FromContext(ctx).Info("chained")

	got := out.String()
	assert.Contains(t, got, "service=api")
	assert.Contains(t, got, "version=v2")

	svcPos := bytes.Index([]byte(got), []byte("service=api"))
	verPos := bytes.Index([]byte(got), []byte("version=v2"))
	assert.Less(t, svcPos, verPos, "earlier WithContext fields should appear first")
}

func TestWithContextEmptyFields(t *testing.T) {
	var out testOutput
	log := New(WithOutput(&out), WithFormatter(text.New()))

	ctx := NewContext(context.Background(), log)
	ctx2 := WithContext(ctx) // no fields

	// Should return a context with the same logger (no crash, no extra output).
	FromContext(ctx2).Info("no extra fields")
	assert.Contains(t, out.String(), "no extra fields")
}
