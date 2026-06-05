package interfaces

import (
	"fmt"
	"time"
)

// FieldType identifies the type of value stored in a Field.
type FieldType uint8

const (
	StringType FieldType = iota
	IntType
	Int64Type
	Float64Type
	BoolType
	DurationType
	TimeType
	ErrorType
	AnyType
)

// Field is a typed key-value pair attached to a log entry.
type Field struct {
	Key   string
	Type  FieldType
	Value any
}

// String returns a Field with a string value.
func String(key, val string) Field {
	return Field{Key: key, Type: StringType, Value: val}
}

// Int returns a Field with an int value.
func Int(key string, val int) Field {
	return Field{Key: key, Type: IntType, Value: val}
}

// Int64 returns a Field with an int64 value.
func Int64(key string, val int64) Field {
	return Field{Key: key, Type: Int64Type, Value: val}
}

// Float64 returns a Field with a float64 value.
func Float64(key string, val float64) Field {
	return Field{Key: key, Type: Float64Type, Value: val}
}

// Bool returns a Field with a bool value.
func Bool(key string, val bool) Field {
	return Field{Key: key, Type: BoolType, Value: val}
}

// Duration returns a Field with a time.Duration value.
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: DurationType, Value: val}
}

// Time returns a Field with a time.Time value.
func Time(key string, val time.Time) Field {
	return Field{Key: key, Type: TimeType, Value: val}
}

// Err returns a Field that captures an error under the key "error".
// If err is nil, the value is stored as the string "<nil>".
func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Type: StringType, Value: "<nil>"}
	}
	return Field{Key: "error", Type: ErrorType, Value: err}
}

// Any returns a Field with an arbitrary value.
func Any(key string, val any) Field {
	return Field{Key: key, Type: AnyType, Value: val}
}

// String returns a human-readable representation of the field for debugging.
func (f Field) String() string {
	return fmt.Sprintf("%s=%v", f.Key, f.Value)
}
