// Package ansi provides utilities for ANSI escape code handling.
// Strip removes colour/style sequences from a string; HasColor detects
// whether a writer supports ANSI colour output.
package ansi

import (
	"io"
	"os"
	"regexp"

	"github.com/mattn/go-isatty"
)

// ansiEscape matches ANSI CSI escape sequences (colours, cursor movement, etc.).
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

// Strip removes all ANSI escape sequences from s, returning plain text.
func Strip(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// HasColor reports whether w appears to support ANSI colour output.
// Returns true when w is an *os.File backed by a terminal (or Cygwin PTY).
func HasColor(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
