package level

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{Trace, "TRACE"},
		{Debug, "DEBUG"},
		{Info, "INFO"},
		{Warn, "WARN"},
		{Error, "ERROR"},
		{Fatal, "FATAL"},
		{Panic, "PANIC"},
		{Level(99), "LEVEL(99)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.level.String())
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  Level
	}{
		{"trace", Trace},
		{"TRACE", Trace},
		{"Trace", Trace},
		{"debug", Debug},
		{"DEBUG", Debug},
		{"Debug", Debug},
		{"info", Info},
		{"INFO", Info},
		{"Info", Info},
		{"warn", Warn},
		{"WARN", Warn},
		{"Warn", Warn},
		{"error", Error},
		{"ERROR", Error},
		{"Error", Error},
		{"fatal", Fatal},
		{"FATAL", Fatal},
		{"Fatal", Fatal},
		{"panic", Panic},
		{"PANIC", Panic},
		{"Panic", Panic},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			require.NoError(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Parse(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParseInvalid(t *testing.T) {
	invalid := []string{"", "verbose", "warning", "critical", "log", "0"}
	for _, s := range invalid {
		t.Run(s, func(t *testing.T) {
			_, err := Parse(s)
			assert.Error(t, err)
			assert.ErrorContains(t, err, s)
		})
	}
}

func TestMustParse(t *testing.T) {
	got := MustParse("info")
	assert.Equal(t, Info, got)
}

func TestMustParsePanics(t *testing.T) {
	assert.Panics(t, func() { MustParse("invalid") })
}

func TestLevelOrdering(t *testing.T) {
	ordered := []Level{Trace, Debug, Info, Warn, Error, Fatal, Panic}
	for i := 1; i < len(ordered); i++ {
		assert.Less(t, ordered[i-1], ordered[i],
			"expected %v < %v", ordered[i-1], ordered[i])
	}
}
