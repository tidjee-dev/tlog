// examples/basic — minimal tlog usage.
//
// Demonstrates creating a logger, logging at every level, and adding fields.
//
// Run:
//
//	go run ./examples/basic/
package main

import (
	"errors"

	"github.com/tidjee-dev/tlog"
	"github.com/tidjee-dev/tlog/level"
)

func main() {
	// Create a logger that writes to stdout.
	// WithConsole auto-selects the pretty formatter on a TTY and plain text otherwise.
	log := tlog.New(
		tlog.WithConsole(),
		tlog.WithLevel(level.Trace),
	)
	defer log.Close()

	log.Trace("tracing internals")
	log.Debug("starting up", tlog.String("env", "development"))
	log.Info("server started", tlog.String("host", "localhost"), tlog.Int("port", 8080))
	log.Warn("high memory usage", tlog.Float64("percent", 87.4))
	log.Error("request failed", tlog.Err(errors.New("connection refused")))
}
