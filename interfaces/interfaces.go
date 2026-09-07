package interfaces

// Formatter converts a log Entry into bytes for output.
// Implementations must be safe for concurrent use; Format is called
// inline on each caller's goroutine.
type Formatter interface {
	Format(entry Entry) ([]byte, error)
}

// Output writes formatted log bytes to a destination.
// Implementations must be safe for concurrent use: Write may race with
// Write, and (once the logger is closing) with Close. Outputs that buffer
// data may additionally implement Sync() error, which Fatal calls
// best-effort before os.Exit.
type Output interface {
	Write(p []byte) error
	Close() error
}

// Encoder is a placeholder for binary/streaming encoders (M2).
type Encoder interface {
	Encode(entry Entry) error
}

// Style is a placeholder for theme abstraction (M3).
type Style interface {
	Apply(s string) string
}
