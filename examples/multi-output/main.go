// examples/multi-output — console and file output simultaneously.
//
// Every log entry is written to both stdout (human-readable) and a JSON file
// (machine-readable). This pattern is common in production services: operators
// tail the console while log aggregators ingest the file.
//
// Run:
//
//	go run ./examples/multi-output/
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tidjee-dev/tlog"
	"github.com/tidjee-dev/tlog/formatter/text"
	"github.com/tidjee-dev/tlog/outputs/console"
)

func main() {
	logPath := filepath.Join(os.TempDir(), "tlog-multi-output-example.log")

	log := tlog.New(
		// Human-readable console (pretty on TTY, plain text otherwise).
		tlog.WithConsole(),
		// Machine-readable JSON appended to a file.
		tlog.WithFile(logPath),
		tlog.WithJSON(),
	)
	defer log.Close()

	log.Info("application started", tlog.String("version", "1.0.0"))
	log.Info("database connected", tlog.String("host", "db.internal"), tlog.Int("port", 5432))
	log.Warn("cache miss rate high", tlog.Float64("miss_rate", 0.42))
	log.Error("failed to send email", tlog.String("to", "user@example.com"))

	fmt.Fprintf(os.Stderr, "\nJSON log written to: %s\n", logPath)

	// Show the file contents on stdout.
	raw, err := os.ReadFile(filepath.Clean(logPath))
	if err == nil {
		fmt.Printf("\n--- %s ---\n%s", logPath, raw)
	}

	// A second logger reusing the same file — entries are appended.
	log2 := tlog.New(
		tlog.WithConsole(console.WithWriter(os.Stdout), console.WithTTY(false)),
		tlog.WithFormatter(text.New()),
		tlog.WithFile(logPath),
		tlog.WithJSON(),
	)
	defer log2.Close()

	log2.Info("second logger appends to same file")
}
