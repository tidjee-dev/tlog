package styles_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tidjee-dev/tlog/styles"
)

var allThemes = []struct {
	name string
	fn   func() styles.Styles
}{
	{"Default", styles.Default},
	{"Minimal", styles.Minimal},
	{"Monochrome", styles.Monochrome},
	{"NoColor", styles.NoColor},
	{"Dev", styles.Dev},
	{"Production", styles.Production},
	{"Badgy", styles.Badgy},
	{"HTTP", styles.HTTP},
	{"BadgyHTTP", styles.BadgyHTTP},
}

// TestThemes_FieldsNonNil verifies every field of each theme is initialised
// (i.e. the style constructors ran without panic).
func TestThemes_FieldsNonNil(t *testing.T) {
	for _, tt := range allThemes {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.fn()
			require.NotNil(t, s.Timestamp)
			require.NotNil(t, s.LevelTrace)
			require.NotNil(t, s.LevelDebug)
			require.NotNil(t, s.LevelInfo)
			require.NotNil(t, s.LevelWarn)
			require.NotNil(t, s.LevelError)
			require.NotNil(t, s.LevelFatal)
			require.NotNil(t, s.LevelPanic)
			require.NotNil(t, s.Message)
			require.NotNil(t, s.Key)
			require.NotNil(t, s.Value)
			require.NotNil(t, s.Caller)
		})
	}
}

// TestThemes_RenderNonEmpty verifies that every theme renders a non-empty
// string for a representative input.
func TestThemes_RenderNonEmpty(t *testing.T) {
	for _, tt := range allThemes {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.fn()
			assert.NotEmpty(t, s.LevelInfo.Render("INFO"))
			assert.NotEmpty(t, s.Message.Render("hello"))
			assert.NotEmpty(t, s.Key.Render("key"))
			assert.NotEmpty(t, s.Value.Render("value"))
		})
	}
}

// TestNoColor_NoANSI verifies that the NoColor theme never emits ANSI codes.
func TestNoColor_NoANSI(t *testing.T) {
	s := styles.NoColor()
	inputs := []string{"INFO", "message text", "key", "value", "file.go:42"}
	styles_ := []struct {
		name   string
		render func(...string) string
	}{
		{"LevelInfo", s.LevelInfo.Render},
		{"Message", s.Message.Render},
		{"Key", s.Key.Render},
		{"Value", s.Value.Render},
		{"Caller", s.Caller.Render},
	}
	for _, st := range styles_ {
		for _, input := range inputs {
			got := st.render(input)
			assert.Equal(t, input, got, "theme=NoColor style=%s input=%q", st.name, input)
		}
	}
}

// TestThemes_Idempotent verifies that calling a theme constructor twice returns
// independent (non-aliased) Styles values.
func TestThemes_Idempotent(t *testing.T) {
	for _, tt := range allThemes {
		t.Run(tt.name, func(t *testing.T) {
			s1 := tt.fn()
			s2 := tt.fn()
			// Record s2's bold state before mutating s1.
			originalBold := s2.Message.GetBold()
			// Toggle bold on s1 — must not affect s2.
			s1.Message = s1.Message.Bold(!originalBold)
			assert.Equal(t, originalBold, s2.Message.GetBold(),
				"mutating s1 should not affect s2 for theme %s", tt.name)
		})
	}
}
