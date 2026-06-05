package interfaces

// Formatter converts a log Entry into bytes for output.
type Formatter interface {
	Format(entry Entry) ([]byte, error)
}

// Output writes formatted log bytes to a destination.
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
