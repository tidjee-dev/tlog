// examples/structured — all tlog field types.
//
// Shows every typed field constructor: String, Int, Int64, Float64, Bool,
// Duration, Time, Err, and Any.
//
// Run:
//
//	go run ./examples/structured/
package main

import (
	"errors"
	"time"

	"github.com/tidjee-dev/tlog"
)

func main() {
	log := tlog.New(tlog.WithConsole())
	defer log.Close()

	log.Info("typed fields",
		tlog.String("service", "payments"),
		tlog.Int("status_code", 200),
		tlog.Int64("request_id", 9876543210),
		tlog.Float64("latency_ms", 12.847),
		tlog.Bool("cache_hit", true),
		tlog.Duration("elapsed", 128*time.Millisecond),
		tlog.Time("started_at", time.Now()),
		tlog.Err(errors.New("upstream timeout")),
		tlog.Any("metadata", map[string]string{"region": "eu-west-1"}),
	)

	// Nil error is safe — rendered as <nil>.
	log.Info("nil error is safe", tlog.Err(nil))

	// Child logger inherits fields.
	req := log.With(
		tlog.String("request_id", "abc-123"),
		tlog.String("user_id", "u_789"),
	)
	req.Info("request received", tlog.String("method", "POST"), tlog.String("path", "/checkout"))
	req.Warn("slow query", tlog.Duration("query_time", 450*time.Millisecond))
}
