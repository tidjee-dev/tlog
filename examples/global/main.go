// examples/global — package-level global logger.
//
// tlog provides an optional package-level default logger. This is convenient
// for small programs and scripts where passing a logger explicitly is
// unnecessary overhead. For library code, prefer explicit logger parameters.
//
// Patterns demonstrated:
//   - SetDefault to install a configured logger as the package default.
//   - Package-level Trace/Debug/Info/Warn/Error convenience functions.
//   - Replacing the default at runtime (e.g. after flag parsing).
//   - The initial no-op default: safe to call before SetDefault.
//
// Run:
//
//	go run ./examples/global/
package main

import (
	"github.com/tidjee-dev/tlog"
	"github.com/tidjee-dev/tlog/level"
)

func main() {
	// Before SetDefault: calls are silently dropped — no panic, no output.
	tlog.Info("this is silently dropped (no-op default)")

	// Install a configured logger as the package-level default.
	tlog.SetDefault(tlog.New(
		tlog.WithConsole(),
		tlog.WithLevel(level.Debug),
	).With(
		tlog.String("app", "my-service"),
	))

	// Package-level helpers now use the configured logger.
	tlog.Debug("debug info", tlog.String("key", "value"))
	tlog.Info("server ready", tlog.Int("port", 9090))
	tlog.Warn("deprecated endpoint called", tlog.String("path", "/v1/legacy"))
	tlog.Error("upstream unavailable", tlog.String("host", "cache.internal"))

	// Replace the default mid-program (e.g. after loading config).
	// The swap is atomic — safe to call from any goroutine.
	tlog.SetDefault(tlog.New(
		tlog.WithConsole(),
		tlog.WithJSON(),
		tlog.WithLevel(level.Info),
	).With(
		tlog.String("app", "my-service"),
		tlog.String("format", "json"),
	))

	tlog.Info("now logging as JSON")
	tlog.Warn("level raised to Info — debug suppressed")
	tlog.Debug("this is suppressed by the Info minimum level")

	// Access the current default logger directly.
	log := tlog.Default()
	log.Info("retrieved via Default()", tlog.Bool("same_logger", true))
	defer log.Close()
}
