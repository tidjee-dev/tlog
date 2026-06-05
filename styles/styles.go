// Package styles defines the Styles struct and built-in themes for tlog's
// pretty console formatter. Each theme is a set of lipgloss.Style values
// applied to the distinct visual regions of a log line.
package styles

import "github.com/charmbracelet/lipgloss"

// Styles holds lipgloss rendering styles for every visual region of a log line.
type Styles struct {
	Timestamp  lipgloss.Style
	LevelTrace lipgloss.Style
	LevelDebug lipgloss.Style
	LevelInfo  lipgloss.Style
	LevelWarn  lipgloss.Style
	LevelError lipgloss.Style
	LevelFatal lipgloss.Style
	LevelPanic lipgloss.Style
	Message    lipgloss.Style
	Key        lipgloss.Style
	Value      lipgloss.Style
	Caller     lipgloss.Style
}

// Dev is a high-contrast, verbose theme for active development. [Default]
func Dev() Styles {
	return Styles{
		Timestamp:  lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		LevelTrace: lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Italic(true),
		LevelDebug: lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
		LevelInfo:  lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
		LevelWarn:  lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true),
		LevelError: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		LevelFatal: lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Bold(true).Underline(true),
		LevelPanic: lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Bold(true).Underline(true),
		Message:    lipgloss.NewStyle().Bold(true),
		Key:        lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		Value:      lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		Caller:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
	}
}

// // Default = "Dev".
// func Default() Styles {
// 	return Dev()
// }

// Minimal is a subdued, low-noise theme for developers who prefer quiet output.
func Minimal() Styles {
	return Styles{
		Timestamp:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		LevelTrace: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		LevelDebug: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		LevelInfo:  lipgloss.NewStyle().Foreground(lipgloss.Color("248")),
		LevelWarn:  lipgloss.NewStyle().Foreground(lipgloss.Color("248")).Bold(true),
		LevelError: lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true),
		LevelFatal: lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true),
		LevelPanic: lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true),
		Message:    lipgloss.NewStyle(),
		Key:        lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Value:      lipgloss.NewStyle(),
		Caller:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
}

// Monochrome uses no color; bold and italic distinguish log regions.
func Monochrome() Styles {
	return Styles{
		Timestamp:  lipgloss.NewStyle(),
		LevelTrace: lipgloss.NewStyle(),
		LevelDebug: lipgloss.NewStyle(),
		LevelInfo:  lipgloss.NewStyle().Bold(true),
		LevelWarn:  lipgloss.NewStyle().Bold(true),
		LevelError: lipgloss.NewStyle().Bold(true),
		LevelFatal: lipgloss.NewStyle().Bold(true).Underline(true),
		LevelPanic: lipgloss.NewStyle().Bold(true).Underline(true),
		Message:    lipgloss.NewStyle(),
		Key:        lipgloss.NewStyle().Italic(true),
		Value:      lipgloss.NewStyle(),
		Caller:     lipgloss.NewStyle().Italic(true),
	}
}

// NoColor is a fully stripped theme — no ANSI codes whatsoever.
// Intended for CI environments and non-TTY destinations.
func NoColor() Styles {
	plain := lipgloss.NewStyle()
	return Styles{
		Timestamp:  plain,
		LevelTrace: plain,
		LevelDebug: plain,
		LevelInfo:  plain,
		LevelWarn:  plain,
		LevelError: plain,
		LevelFatal: plain,
		LevelPanic: plain,
		Message:    plain,
		Key:        plain,
		Value:      plain,
		Caller:     plain,
	}
}

// HTTP is a theme tuned for web-server access logging. Cyan keys make
// method/path/status/latency easy to scan; green INFO signals success,
// yellow WARN flags slow or redirected requests, and red ERROR stands out
// for 5xx failures. Values are rendered in warm yellow for quick reads.
func HTTP() Styles {
	return Styles{
		Timestamp:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		LevelTrace: lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true),
		LevelDebug: lipgloss.NewStyle().Foreground(lipgloss.Color("75")),
		LevelInfo:  lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true),
		LevelWarn:  lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true),
		LevelError: lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true),
		LevelFatal: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Underline(true),
		LevelPanic: lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true).Underline(true),
		Message:    lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		Key:        lipgloss.NewStyle().Foreground(lipgloss.Color("81")),
		Value:      lipgloss.NewStyle().Foreground(lipgloss.Color("222")),
		Caller:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
	}
}

// BadgyHTTP combines badge-style level rendering with HTTP-oriented colours.
// Levels appear as filled pills; cyan keys and warm-yellow values make
// method/path/status/latency instantly readable.
func BadgyHTTP() Styles {
	badge := func(bg, fg string) lipgloss.Style {
		return lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(fg)).
			PaddingLeft(1).PaddingRight(1)
	}
	return Styles{
		Timestamp:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		LevelTrace: badge("239", "250"),
		LevelDebug: badge("75", "0"),
		LevelInfo:  badge("36", "0").Bold(true),
		LevelWarn:  badge("220", "0").Bold(true),
		LevelError: badge("203", "255").Bold(true),
		LevelFatal: badge("124", "255").Bold(true).Underline(true),
		LevelPanic: badge("201", "0").Bold(true).Underline(true),
		Message:    lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		Key:        lipgloss.NewStyle().Foreground(lipgloss.Color("81")),
		Value:      lipgloss.NewStyle().Foreground(lipgloss.Color("222")),
		Caller:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
	}
}

// Badgy is a vivid theme where each level is rendered as a filled badge
// (coloured background + contrasting foreground) for instant visual triage.
func Badgy() Styles {
	badge := func(bg, fg string) lipgloss.Style {
		return lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(fg)).
			PaddingLeft(1).PaddingRight(1)
	}
	return Styles{
		Timestamp:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		LevelTrace: badge("239", "255"),
		LevelDebug: badge("27", "255"),
		LevelInfo:  badge("36", "0"),
		LevelWarn:  badge("214", "0"),
		LevelError: badge("196", "255").Bold(true),
		LevelFatal: badge("88", "255").Bold(true).Underline(true),
		LevelPanic: badge("200", "0").Bold(true).Underline(true),
		Message:    lipgloss.NewStyle(),
		Key:        lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		Value:      lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Caller:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
	}
}

// Production is a subdued theme optimised for ops/on-call viewing.
func Production() Styles {
	return Styles{
		Timestamp:  lipgloss.NewStyle().Foreground(lipgloss.Color("242")),
		LevelTrace: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		LevelDebug: lipgloss.NewStyle().Foreground(lipgloss.Color("242")),
		LevelInfo:  lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		LevelWarn:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		LevelError: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		LevelFatal: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Underline(true),
		LevelPanic: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Underline(true),
		Message:    lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Key:        lipgloss.NewStyle().Foreground(lipgloss.Color("242")),
		Value:      lipgloss.NewStyle().Foreground(lipgloss.Color("248")),
		Caller:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
	}
}
