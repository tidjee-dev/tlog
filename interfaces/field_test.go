package interfaces

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestFieldConstructors(t *testing.T) {
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	dur := 322 * time.Millisecond
	err := errors.New("something went wrong")

	tests := []struct {
		name string
		got  Field
		want Field
	}{
		{
			name: "String",
			got:  String("host", "localhost"),
			want: Field{Key: "host", Type: StringType, Value: "localhost"},
		},
		{
			name: "Int",
			got:  Int("port", 8080),
			want: Field{Key: "port", Type: IntType, Value: 8080},
		},
		{
			name: "Int64",
			got:  Int64("count", int64(1_000_000)),
			want: Field{Key: "count", Type: Int64Type, Value: int64(1_000_000)},
		},
		{
			name: "Float64",
			got:  Float64("ratio", 0.95),
			want: Field{Key: "ratio", Type: Float64Type, Value: 0.95},
		},
		{
			name: "Bool",
			got:  Bool("enabled", true),
			want: Field{Key: "enabled", Type: BoolType, Value: true},
		},
		{
			name: "Duration",
			got:  Duration("latency", dur),
			want: Field{Key: "latency", Type: DurationType, Value: dur},
		},
		{
			name: "Time",
			got:  Time("started_at", now),
			want: Field{Key: "started_at", Type: TimeType, Value: now},
		},
		{
			name: "Err",
			got:  Err(err),
			want: Field{Key: "error", Type: ErrorType, Value: err},
		},
		{
			name: "Any",
			got:  Any("meta", map[string]int{"a": 1}),
			want: Field{Key: "meta", Type: AnyType, Value: map[string]int{"a": 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, tt.got, cmp.Comparer(func(a, b error) bool {
				return a.Error() == b.Error()
			})); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestErrNil(t *testing.T) {
	f := Err(nil)
	assert.Equal(t, "error", f.Key)
	assert.Equal(t, StringType, f.Type)
	assert.Equal(t, "<nil>", f.Value)
}

func TestFieldString(t *testing.T) {
	f := String("host", "localhost")
	assert.Equal(t, "host=localhost", f.String())
}
