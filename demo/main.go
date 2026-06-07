package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/tidjee-dev/tlog"
	"github.com/tidjee-dev/tlog/outputs/console"
)

func main() {
	prettyLog := tlog.New(
		tlog.WithConsole(),
	)

	jsonLog := tlog.New(
		tlog.WithConsole(),
		tlog.WithJSON(),
	)

	textLog := tlog.New(
		tlog.WithConsole(
			console.WithTTY(false),
		),
	)

	fileLog := tlog.New(
		tlog.WithFile("tlog-dev-demo2", "demo.log"),
	)

	fileJsonLog := tlog.New(
		tlog.WithFile("tlog-dev-demo2", "demo.jsonl"),
		tlog.WithJSON(),
	)

	// Simulate some work
	start := time.Now()
	// time.Sleep(1 * time.Nanosecond)
	elapsed := time.Since(start)
	fmt.Printf("Simulated work duration: %s\n", elapsed)

	// Random duration for demonstration (0-1day)
	ttl := time.Duration(rand.Intn(86_400_000)) * time.Millisecond // #nosec G404
	fmt.Printf("Simulated duration: %s\n", ttl)

	fields := []tlog.Field{
		tlog.Int("id", 123),
		tlog.String("name", "John Doe"),
		tlog.Bool("active", true),
		tlog.Float64("score", 99.5),
		tlog.Duration("duration_elapsed", elapsed),
		tlog.Duration("duration_ttl", ttl),
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
	fmt.Println()
	textLog.Info("user created", fields...)
	fileLog.Info("user created", fields...)
	fileJsonLog.Info("user created", fields...)
}
