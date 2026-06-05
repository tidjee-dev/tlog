package ansi_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidjee-dev/tlog/pkg/ansi"
)

func TestStrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no_escape",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "reset",
			input: "\x1b[0mhello\x1b[0m",
			want:  "hello",
		},
		{
			name:  "colour",
			input: "\x1b[31mERROR\x1b[0m something went wrong",
			want:  "ERROR something went wrong",
		},
		{
			name:  "bold_italic",
			input: "\x1b[1m\x1b[3mtext\x1b[0m",
			want:  "text",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "only_escapes",
			input: "\x1b[31m\x1b[0m",
			want:  "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ansi.Strip(tc.input))
		})
	}
}

func TestHasColor_NonFile(t *testing.T) {
	// bytes.Buffer is not an *os.File — HasColor must return false.
	assert.False(t, ansi.HasColor(&bytes.Buffer{}))
}
