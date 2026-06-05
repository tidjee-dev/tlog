// examples/custom-theme — lipgloss style customisation.
//
// Shows how to:
//   - Select a built-in theme with WithTheme.
//   - Override individual style fields with WithStyles.
//   - Build a fully custom theme from scratch.
//
// Run:
//
//	go run ./examples/custom-theme/
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"

	"github.com/tidjee-dev/tlog"
	"github.com/tidjee-dev/tlog/level"
	"github.com/tidjee-dev/tlog/styles"
)

func separator(label string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#fff")).Bold(true)
	_, _ = fmt.Fprintf(os.Stdout, "\n%s\n", style.Render(fmt.Sprintf("── %s ──", label)))
}

func demo(log *tlog.Logger) {
	log.Info("request handled", tlog.String("path", "/api/v1/users"), tlog.Int("status", 200))
	log.Warn("rate limit approaching", tlog.Int("remaining", 10))
	log.Error("upstream failed", tlog.Err(errors.New("connection reset")))
}

func main() {
	// ── Built-in themes ───────────────────────────────────────────────────────

	for _, tc := range []struct {
		name  string
		theme styles.Styles
	}{
		{"Dev (default)", styles.Dev()},
		{"Minimal", styles.Minimal()},
		{"Monochrome", styles.Monochrome()},
		{"Production", styles.Production()},
		{"Badgy", styles.Badgy()},
		{"HTTP", styles.HTTP()},
		{"BadgyHTTP", styles.BadgyHTTP()},
		{"NoColor", styles.NoColor()},
	} {
		separator(tc.name)
		log := tlog.New(
			tlog.WithConsole(),
			tlog.WithTheme(tc.theme),
			tlog.WithLevel(level.Trace),
		)
		demo(log)
		log.Close()
	}

	// ── Partial override with WithStyles ──────────────────────────────────────
	separator("Partial override (Dev base, red errors, italic keys)")
	log := tlog.New(
		tlog.WithConsole(),
		tlog.WithTheme(styles.Dev()),
		tlog.WithStyles(func(s *tlog.Styles) {
			s.LevelError = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Bold(true).
				Reverse(true)
			s.Key = lipgloss.NewStyle().
				Foreground(lipgloss.Color("45")).
				Italic(true)
		}),
		tlog.WithLevel(level.Trace),
	)
	demo(log)
	log.Close()

	// ── Fully custom theme ────────────────────────────────────────────────────
	separator("Fully custom theme")
	custom := styles.Styles{
		Timestamp:  lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		LevelTrace: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		LevelDebug: lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
		LevelInfo:  lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true),
		LevelWarn:  lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true),
		LevelError: lipgloss.NewStyle().Background(lipgloss.Color("196")).Foreground(lipgloss.Color("255")).Bold(true),
		LevelFatal: lipgloss.NewStyle().Background(lipgloss.Color("88")).Foreground(lipgloss.Color("255")).Bold(true),
		LevelPanic: lipgloss.NewStyle().Background(lipgloss.Color("88")).Foreground(lipgloss.Color("255")).Bold(true),
		Message:    lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		Key:        lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		Value:      lipgloss.NewStyle().Foreground(lipgloss.Color("220")),
		Caller:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
	}
	log2 := tlog.New(
		tlog.WithConsole(),
		tlog.WithTheme(custom),
		tlog.WithLevel(level.Trace),
		tlog.WithCaller(),
	)
	demo(log2)
	log2.Close()
}
