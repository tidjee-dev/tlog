package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/tidjee-dev/tlog"
)

func main() {
	start := time.Now()
	time.Sleep(150 * time.Millisecond) // Simulate some work

	prettyLog := tlog.New(
		tlog.WithConsole(),
	)

	jsonLog := tlog.New(
		tlog.WithConsole(),
		tlog.WithJSON(),
	)

	fields := []tlog.Field{
		tlog.Int("id", 123),
		tlog.String("name", "John Doe"),
		tlog.Bool("active", true),
		tlog.Float64("score", 99.5),
		tlog.Duration("duration", time.Since(start)),
		tlog.Err(errors.New("something failed")),
		tlog.Any("any_string", "raw-any"),
		tlog.Any("any_struct", map[string]any{
			"role": "admin",
			"age":  30,
		}),
	}

	prettyLog.Info("user created", fields...)
	fmt.Println()
	jsonLog.Info("user created", fields...)
}
