// Package discard provides an Output that silently drops all log entries.
// It is useful in tests and benchmarks where output is not needed.
package discard

// Discard is an Output that drops every write and always returns nil.
type Discard struct{}

// New returns a Discard output.
func New() *Discard { return &Discard{} }

// Write implements interfaces.Output. It discards p and returns nil.
func (*Discard) Write(_ []byte) error { return nil }

// Close implements interfaces.Output. It is a no-op and returns nil.
func (*Discard) Close() error { return nil }
