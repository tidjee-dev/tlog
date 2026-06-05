// examples/json-mode — structured JSON logging for production.
//
// In production you typically want:
//   - JSON output for log aggregators (Loki, Datadog, Splunk, …).
//   - RFC 3339 timestamps (default).
//   - All context fields attached at logger construction time.
//   - Caller info for quick source location.
//   - An error handler that records logger failures without crashing.
//
// Run:
//
//	go run ./examples/json-mode/
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	jsonfmt "github.com/tidjee-dev/tlog/formatter/json"

	"github.com/tidjee-dev/tlog"
)

func main() {
	// Production logger: JSON to stdout, caller info, service-level fields.
	log := tlog.New(
		tlog.WithConsole(),
		tlog.WithJSON(jsonfmt.WithTimeFormat(time.RFC3339Nano)),
		tlog.WithCaller(),
		tlog.WithErrorHandler(func(err error) {
			fmt.Fprintf(os.Stderr, "tlog error: %v\n", err)
		}),
	).With(
		tlog.String("service", "payment-service"),
		tlog.String("version", "2.3.1"),
		tlog.String("env", "production"),
	)
	defer log.Close()

	log.Info("service started", tlog.Int("pid", os.Getpid()))

	// Simulate request handling with per-request child loggers.
	handleRequest(log, "req-001", "POST", "/api/v1/charge")
	handleRequest(log, "req-002", "GET", "/api/v1/status")
}

func handleRequest(base *tlog.Logger, id, method, path string) {
	log := base.With(
		tlog.String("request_id", id),
		tlog.String("method", method),
		tlog.String("path", path),
	)

	log.Info("request received")

	// Simulate some work.
	start := time.Now()
	err := processPayment(id)
	elapsed := time.Since(start)

	if err != nil {
		log.Error("request failed",
			tlog.Err(err),
			tlog.Duration("elapsed", elapsed),
			tlog.Int("status", 500),
		)
		return
	}

	log.Info("request completed",
		tlog.Duration("elapsed", elapsed),
		tlog.Int("status", 200),
	)
}

func processPayment(id string) error {
	if id == "req-001" {
		return nil
	}
	return errors.New("card declined")
}
